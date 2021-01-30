// Copyright 2020-2021 Team RelicFrog
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// ʕ◔ϖ◔ʔ
//

package main

import (
	rfpb "api_usr_role/server/proto"
	"context"
	"fmt"
	rftlp "github.com/RelicFrog/go-lib-pub-tlp"
	"github.com/golang/protobuf/ptypes"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	rfpbh "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	metaServiceVersion       = "1.0.0"
	metaServiceName          = "grpc_usr_role"
	metaMongoDbCollectionTbl = "user_roles"
)

var (
	log *logrus.Logger
	metaMongoDbUsr  string
	metaMongoDbPwd  string
	metaMongoDbLnk  string
	metaMongoDbPDB  string
    metaMongoDbClient *mongo.Client
	metaMongoDbCollection *mongo.Collection
	metaMongoDbContext = context.Background()
	metaEnvFile = ".env"
	metaServicePort string
	extraLatency time.Duration
    err error
)

type UserRoleServiceServer struct {}
type UserRole struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	MetaName        string             `bson:"meta_name"`
	MetaDescription string             `bson:"meta_description"`
	MetaAppHandle   string             `bson:"meta_app_handle"`
	MetaAppIcon     string             `bson:"meta_app_icon"`
	MetaAppColorHex string             `bson:"meta_app_color_hex"`
	IsLocked        bool               `bson:"is_locked"`
	IsDeleted       bool               `bson:"is_deleted"`
	IsFixture       bool               `bson:"is_fixture"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

//
// -- gRPC Server Stack 0/n :: Server Logic
//

func init() {

	//
	// -- define logging mechanics --
	//

	log = logrus.New()
	log.Level = logrus.DebugLevel
	if metaDebugMode := _getDotEnvVariable("DISABLE_DEBUG",metaEnvFile); metaDebugMode == "1" {
		log.Level = logrus.ErrorLevel
	}

	log.Formatter = &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyMsg:   "message",
		},
		TimestampFormat: time.RFC3339Nano,
	};  log.Out = os.Stdout

	//
	// -- define config bound mechanics (using .env) --
	//

	if metaTracerDisabled := _getDotEnvVariable("DISABLE_TRACING",metaEnvFile); metaTracerDisabled == "0" {
		log.Infof("%s: tracing enabled.",metaServiceName)
		go rftlp.InitTracing(metaServiceName, _getDotEnvVariable("JAEGER_SERVICE_ADDR",metaEnvFile), log)
	}

	if metaProfilerDisabled := _getDotEnvVariable("DISABLE_PROFILER",metaEnvFile); metaProfilerDisabled == "0" {
		log.Infof("%s: profiling enabled.",metaServiceName)
		go rftlp.InitProfiling(metaServiceName, metaServiceVersion, log)
	}

	if metaMongoDbUsr = _getDotEnvVariable("DB_MONGO_USR",metaEnvFile); metaMongoDbUsr == "" {
		log.Fatalf("%s: mongoDB-Service-User not set <exit>",metaServiceName)
	}

	if metaMongoDbPwd = _getDotEnvVariable("DB_MONGO_PWD",metaEnvFile); metaMongoDbPwd == "" {
		log.Fatalf("%s: mongoDB-Password not set <exit>",metaServiceName)
	}

	if metaMongoDbPDB = _getDotEnvVariable("DB_MONGO_PDB",metaEnvFile); metaMongoDbPDB == "" {
		log.Fatalf("%s: mongoDB primary service db not found <exit>",metaServiceName)
	}

	if metaMongoDbLnk = _getDotEnvVariable("DB_MONGO_LNK",metaEnvFile); metaMongoDbLnk == "" {
		log.Fatalf("%s: mongoDB connection link not found <exit>",metaServiceName)
	}

	if metaServicePort = _getDotEnvVariable("PORT",metaEnvFile); metaServicePort == "" {
		log.Fatalf("%s: service port definition missing <exit>",metaServiceName)
	}
}

func main() {

	log.Infof("%s: start",metaServiceName)

	//
	// -- primary startup sequence --
	//

	mongoDbInitCon()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,syscall.SIGINT,syscall.SIGTERM,syscall.SIGABRT,syscall.SIGUSR1,syscall.SIGUSR2)
	go func() {

		for {

			sig := <-sigs ; log.Printf("%s: received [%s] signal",metaServiceName,sig.String())
			// handle service abort signals (SIGINT/SIGTERM/SIGABRT)
			if sig == syscall.SIGINT ||  sig == syscall.SIGTERM ||  sig == syscall.SIGABRT  {
				log.Infof("%s: handle QUIT signal [%s] ...",metaServiceName,sig.String())
				log.Infof("%s: done",metaServiceName)

				mongoDbCloseCon()
				os.Exit(0)
			}

			// handle (internal) fixture load signal (notify on syscall.USR1)
			if sig == syscall.SIGUSR1 {
				log.Infof("%s: handle seed database signal [%s] ...",metaServiceName,sig.String())

				// if err := mongoDbFixtureCreateIndexes(); err != nil { mongoDbFixtureHandleFatal(err) }
				if err := mongoDbFixtureLoadUserRoles(); err != nil { mongoDbFixtureHandleFatal(err) }
			}

			// handle extra latency signal (notify on syscall.USR2)
			if sig == syscall.SIGUSR2 {
				log.Infof("%s: handle extra latency signal [%s] ...",metaServiceName,sig.String())
				// check inactive state (switch latency flag)
				if  extraLatency == time.Duration(0) {
					extraLatency = time.Duration(rand.Int63n(1750))*time.Millisecond
					log.Infof("%s: extra latency enabled (duration: %v) ...",metaServiceName,extraLatency)
				} else {
					extraLatency = time.Duration(0)
					log.Infof("%s: extra latency disabled (duration: %v) ...",metaServiceName,extraLatency)
				}
			}
		}
	}()

	log.Infof("%s: starting gRPC server at port [%s] ...",metaServiceName,metaServicePort)
	run(metaServicePort); select {}
}

func run(port string) string {

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil { log.Fatal(err) }

	var srv *grpc.Server
	if metaGRPCStatsDisabled := _getDotEnvVariable("DISABLE_STATS",metaEnvFile); metaGRPCStatsDisabled == "0" {
		log.Infof("%s: gRPC stats enabled.",metaServiceName)
		srv = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	} else {
		srv = grpc.NewServer()
	}

	userRoleSVC := &UserRoleServiceServer{}
	rfpb.RegisterUserRoleServiceServer(srv, userRoleSVC)
	rfpbh.RegisterHealthServer(srv, userRoleSVC)
	reflection.Register(srv) // activate reflections

	go srv.Serve(l)

	log.Infof("%s: send SIG.TERM or SIG.INT (CTRL+c) to quit this gRPC endpoint ...",metaServiceName)

	return l.Addr().String()
}

//
// -- gRPC Method Stack 1/n :: HealthCheck(s)
//

func (u UserRoleServiceServer) Check(_ context.Context, _ *rfpbh.HealthCheckRequest) (*rfpbh.HealthCheckResponse, error) {
	return &rfpbh.HealthCheckResponse{Status: rfpbh.HealthCheckResponse_SERVING}, nil
}

func (u UserRoleServiceServer) Watch(_ *rfpbh.HealthCheckRequest, _ rfpbh.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

//
// -- gRPC Method Stack 2/n :: MetaOps
//

func (u UserRoleServiceServer) GetVersion(_ context.Context, _ *rfpb.VersionReq) (*rfpb.VersionRes, error) {

	return &rfpb.VersionRes{Version: fmt.Sprintf("v%s", metaServiceVersion)}, nil
}

func (u UserRoleServiceServer) CreateRole(_ context.Context, req *rfpb.CreateRoleReq) (*rfpb.CreateRoleRes, error) {

	// essentially doing req.GetRole to access the struct with a nil check
	metaRole := req.GetRole()
	metaData := UserRole{
		MetaName: metaRole.GetMetaName(),
		MetaDescription: metaRole.GetMetaDescription(),
		MetaAppColorHex: metaRole.GetMetaAppColorHex(),
		MetaAppIcon: metaRole.GetMetaAppIcon(),
		MetaAppHandle: metaRole.GetMetaAppHandle(),
		IsDeleted: metaRole.GetIsDeleted(),
		IsLocked: metaRole.GetIsLocked(),
		IsFixture: metaRole.GetIsFixture(),
		CreatedAt: time.Now(),
	}

	log.Infof("%s: CreateRole: receive gRPC role-guid: %s",metaServiceName,metaData.MetaName)
	result, err := metaMongoDbCollection.InsertOne(metaMongoDbContext, metaData)
	if err != nil {
		log.Warnf("%s: mongodb: error during gRPC based insert operation of given user-role",metaServiceName)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	metaRole.Id = result.InsertedID.(primitive.ObjectID).Hex()
	log.Infof("%s: CreateRole: persist gRPC oid: %s",metaServiceName,metaRole.Id)

	return &rfpb.CreateRoleRes{ Role: metaRole }, nil
}

func (u UserRoleServiceServer) GetRole(_ context.Context, req *rfpb.GetRoleReq) (*rfpb.GetRoleRes, error) {

	// convert string id (from proto) to mongoDB ObjectId
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert object-id to document-id",metaServiceName)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	log.Infof("%s: GetRole: receive gRPC role-oid: %s",metaServiceName,req.GetId())
	result := metaMongoDbCollection.FindOne(metaMongoDbContext, bson.M{"_id": oid, "is_deleted": false})

	metaRole := UserRole{}
	if err := result.Decode(&metaRole); err != nil {
		log.Warnf("%s: mongodb: unable to find document with object-id %s",metaServiceName,req.GetId())
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Get Role Fail (!) -> Error: %v", err))
	}

	// prepare some core type variables
	tsMetaCreatedAt, _ := ptypes.TimestampProto(metaRole.CreatedAt)
	tsMetaUpdatedAt, _ := ptypes.TimestampProto(metaRole.UpdatedAt)
	// Cast to GetRoleRes type
	response := &rfpb.GetRoleRes{
		Role: &rfpb.UserRole{
			Id: oid.Hex(),
			MetaName: metaRole.MetaName,
			MetaDescription: metaRole.MetaDescription,
			MetaAppHandle: metaRole.MetaAppHandle,
			MetaAppColorHex: metaRole.MetaAppColorHex,
			MetaAppIcon: metaRole.MetaAppIcon,
			CreatedAt: tsMetaCreatedAt,
			UpdatedAt: tsMetaUpdatedAt,
			IsFixture: metaRole.IsFixture,
			IsLocked: metaRole.IsLocked,
		},
	}

	return response, nil
}

func (u UserRoleServiceServer) DeleteRole(_ context.Context, req *rfpb.DeleteRoleReq) (*rfpb.DeleteRoleRes, error) {

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert object-id (%s) to document-id!",metaServiceName,req.GetId())
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	result := metaMongoDbCollection.FindOneAndUpdate(metaMongoDbContext,
		bson.M{"_id": oid, "is_deleted": false},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": time.Now(),
		}},  options.FindOneAndUpdate().SetReturnDocument(1))


	decoded := UserRole{}
	err = result.Decode(&decoded)
	if err != nil {
		log.Warnf("%s: mongodb: unable to find role with supplied ID: %s",metaServiceName,oid)
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Delete Role Fail (!) -> Error: %v", err))
	}

	return &rfpb.DeleteRoleRes{ Success: true }, nil
}

func (u UserRoleServiceServer) UpdateRole(_ context.Context, req *rfpb.UpdateRoleReq) (*rfpb.UpdateRoleRes, error) {

	metaRole := req.GetRole()
	oid, err := primitive.ObjectIDFromHex(metaRole.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert oid (raw=%s), object=%v",metaServiceName,metaRole.GetId(),req)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	res := metaMongoDbCollection.FindOneAndUpdate(metaMongoDbContext,
		bson.M{"_id": oid, "is_deleted": false},
		bson.M{"$set": bson.M{
				"meta_name": metaRole.MetaName,
				"meta_description": metaRole.MetaDescription,
				"meta_app_handle": metaRole.MetaAppHandle,
				"meta_app_icon": metaRole.MetaAppIcon,
			    "meta_app_color_hex": metaRole.MetaAppColorHex,
				"updated_at": time.Now(),
			}},  options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := UserRole{}
	err = res.Decode(&decoded)
	if err != nil {
		log.Warnf("%s: mongodb: unable to find document with oid: %s",metaServiceName,oid)
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Update Role Fail (!) -> Error: %v", err))
	}

	tsUpdatedAt, _ := ptypes.TimestampProto(decoded.UpdatedAt)
	return &rfpb.UpdateRoleRes{
		Role: &rfpb.UserRole{
			Id:       		 decoded.ID.Hex(),
			MetaName: 		 decoded.MetaName,
			MetaDescription: decoded.MetaDescription,
			MetaAppHandle:   decoded.MetaAppHandle,
			MetaAppIcon:  	 decoded.MetaAppIcon,
			MetaAppColorHex: decoded.MetaAppColorHex,
			UpdatedAt:       tsUpdatedAt,
		},
	}, nil
}

func (u UserRoleServiceServer) ListRoles(_ *rfpb.ListRoleReq, stream rfpb.UserRoleService_ListRolesServer) error {

	data := &UserRole{}
	cursor, err := metaMongoDbCollection.Find(metaMongoDbContext, bson.M{"is_deleted": false})
	if err != nil {
		log.Warnf("%s: mongodb: unable to find user-role(s)",metaServiceName)
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	};  defer cursor.Close(metaMongoDbContext)

	for cursor.Next(metaMongoDbContext) {
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}
		// prepare some core type variables
		tsCreatedAt, _ := ptypes.TimestampProto(data.CreatedAt)
		tsUpdatedAt, _ := ptypes.TimestampProto(data.UpdatedAt)
		// if no error is found send user roles via stream
		_ = stream.Send(&rfpb.ListRoleRes{
			Role: &rfpb.UserRole{
				Id:              data.ID.Hex(),
				MetaName: 		 data.MetaName,
				MetaDescription: data.MetaDescription,
				MetaAppHandle:   data.MetaAppHandle,
				MetaAppIcon:  	 data.MetaAppIcon,
				MetaAppColorHex: data.MetaAppColorHex,
				CreatedAt:       tsCreatedAt,
				UpdatedAt:       tsUpdatedAt,
			},
		})
	}

	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unkown cursor error: %v", err))
	}

	return nil
}

//
// -- gRPC MongoDb Stack 3/n :: MongoDbOps
//

func mongoDbFixtureClean() error {

	dr, err := metaMongoDbCollection.DeleteMany(metaMongoDbContext, bson.D{{}})

	if err != nil { return err }
	if dr.DeletedCount > 0 {
		log.Infof("%s: mongodb: delete %v documents in [%s] metaMongoDbCollection",metaServiceName,dr.DeletedCount,metaMongoDbCollection.Name())
	}

	return nil
}

func mongoDbFixtureLoadUserRoles() error {

	if err = mongoDbFixtureClean(); err != nil { return err }

	// -- load [user_invitation_codes] metaMongoDbCollection for [admins] --
	log.Infof("%s: mongodb: generate [admin] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	if _genUserRole("ROLE_ADMIN", "admin", "security", "#CB212D") != nil {
		return err
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [directors (powerUsers)] --
	log.Infof("%s: mongodb: generate [director] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	if _genUserRole("ROLE_DIRECTOR", "director", "verified_user", "#77A548") != nil {
		return err
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [teacher] --
	log.Infof("%s: mongodb: generate [teacher] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	if _genUserRole("ROLE_TEACHER", "teacher", "school", "#38557D") != nil {
		return err
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [viewer] --
	log.Infof("%s: mongodb: generate [viewer] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	if _genUserRole("ROLE_VIEWER", "viewer", "check", "#F5AE3B") != nil {
		return err
	}

	return nil
}

func mongoDbFixtureHandleFatal(err error) {

	mongoDbCloseCon()
	log.Fatal(err)
}

func mongoDbFixtureCreateIndexes() error {

	// create index for meta_valid_from, meta_valid_to and meta_code (unique)
	models := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "meta_name", Value: bsonx.String("text")}},
			Options: options.Index().SetName("meta_name_idx").SetUnique(true),
        },
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err = metaMongoDbCollection.Indexes().CreateMany(metaMongoDbContext, models, opts)
	if err != nil { return err }

	return nil
}

func mongoDbInitCon() {

	clientOptions := options.Client().ApplyURI(metaMongoDbLnk).SetAuth(options.Credential{
		AuthSource: metaMongoDbPDB, Username: metaMongoDbUsr, Password: metaMongoDbPwd,
	})

	// declare Context type object for managing multiple API requests
	metaMongoDbContext, _ := context.WithTimeout(context.Background(), 10*time.Second)
	metaMongoDbClient, err = mongo.Connect(metaMongoDbContext, clientOptions)
	if err != nil { log.Fatal(err) }

	err = metaMongoDbClient.Ping(metaMongoDbContext, nil)
	if err != nil {
		log.Fatal(err)
	}

	metaMongoDbCollection = metaMongoDbClient.Database(metaMongoDbPDB).Collection(metaMongoDbCollectionTbl)

	log.Infof("%s: mongodb: connection opened",metaServiceName)
}

func mongoDbCloseCon() {

	err = metaMongoDbClient.Disconnect(metaMongoDbContext)
	if err != nil { log.Fatal(err) }

	log.Infof("%s: mongodb: connection closed",metaServiceName)
}

//
// -- sidekick stack for MongoDbOps && (some) gRPC helper methods
//

func _genUserRole(roleName string, roleHandle string, roleIcon string, roleColorCodeHex string) error {

	_, err = metaMongoDbCollection.InsertOne(metaMongoDbContext, UserRole{
		IsFixture: true,
		IsLocked: true,
		IsDeleted: false,
		MetaName: roleName,
		MetaDescription: fmt.Sprintf("auto-generated role [%s] using handle [%s]",roleName,roleHandle),
		MetaAppIcon: roleIcon,
		MetaAppColorHex: roleColorCodeHex,
		MetaAppHandle: roleHandle,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}); if err != nil {
		log.Infof("%s: mongodb: add role [%s]-[%s] failed, code may already exists ...",metaServiceName,roleName,roleHandle)
		return err
	}

	log.Infof("%s: mongodb: add role [%s]-[%s]",metaServiceName,roleName,roleHandle)

	return nil
}

func _getDotEnvVariable(key string, file ...string) string {

	envFile := ".env"
	if len(file) > 0 {
		envFile = file[0]
	}

	if err := godotenv.Load(envFile) ; err != nil {
		log.Fatalf("common: error loading local [.%s] file",file)
	}

	return os.Getenv(key)
}
