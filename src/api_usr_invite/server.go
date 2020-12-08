// Copyright 2020 Team RelicFrog
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
	rfpb "api_usr_invite/server/proto"
	"context"
	"fmt"
	rftlp "github.com/RelicFrog/go-lib-pub-tlp"
	"github.com/golang/protobuf/ptypes"
	"github.com/joho/godotenv"
	"github.com/oklog/ulid/v2"
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
	metaServiceName          = "grpc_usr_invite"
	metaMongoDbCollectionTbl = "user_codes"
)

var (
	log *logrus.Logger
	metaMongoDbUsr string
	metaMongoDbPwd string
	metaMongoDbLnk string
	metaMongoDbPDB string
    metaMongoDbClient *mongo.Client
	metaMongoDbCollection *mongo.Collection
	metaMongoDbContext = context.Background()
	metaServicePort string
	extraLatency time.Duration
    err error
)

type UserInviteCodeServiceServer struct {}
type UserInviteCode struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	MetaCode       string             `bson:"meta_code"`
	MetaForAppRole string             `bson:"meta_for_app_role"`
	MetaValidFrom  time.Time          `bson:"meta_valid_from"`
	MetaValidTo    time.Time          `bson:"meta_valid_to"`
	CreatedAt      time.Time          `bson:"created_at"`
	IsFixture      bool               `bson:"is_fixture"`
	IsDeleted      bool               `bson:"is_deleted"`
	IsTest         bool               `bson:"is_test"`
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
	if metaDebugMode := _getDotEnvVariable("DISABLE_DEBUG"); metaDebugMode == "1" {
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

	if metaTracerDisabled := _getDotEnvVariable("DISABLE_TRACING"); metaTracerDisabled == "0" {
		log.Infof("%s: tracing enabled.",metaServiceName)
		go rftlp.InitTracing(metaServiceName, _getDotEnvVariable("JAEGER_SERVICE_ADDR"), log)
	}

	if metaProfilerDisabled := _getDotEnvVariable("DISABLE_PROFILER"); metaProfilerDisabled == "0" {
		log.Infof("%s: profiling enabled.",metaServiceName)
		go rftlp.InitProfiling(metaServiceName, metaServiceVersion, log)
	}

	if metaMongoDbUsr = _getDotEnvVariable("DB_MONGO_USR"); metaMongoDbUsr == "" {
		log.Fatalf("%s: mongoDB-Service-User not set <exit>",metaServiceName)
	}

	if metaMongoDbPwd = _getDotEnvVariable("DB_MONGO_PWD"); metaMongoDbPwd == "" {
		log.Fatalf("%s: mongoDB-Password not set <exit>",metaServiceName)
	}

	if metaMongoDbPDB = _getDotEnvVariable("DB_MONGO_PDB"); metaMongoDbPDB == "" {
		log.Fatalf("%s: mongoDB primary service db not found <exit>",metaServiceName)
	}

	if metaMongoDbLnk = _getDotEnvVariable("DB_MONGO_LNK"); metaMongoDbLnk == "" {
		log.Fatalf("%s: mongoDB connection link not found <exit>",metaServiceName)
	}

	if metaServicePort = _getDotEnvVariable("PORT"); metaServicePort == "" {
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

			// handle fixture load / db seeding signal (notify on syscall.USR1)
			if sig == syscall.SIGUSR1 {
				log.Infof("%s: handle seed database signal [%s] ...",metaServiceName,sig.String())
				if err := mongoDbFixtureCreateIndexes(); err != nil { mongoDbFixtureHandleFatal(err) }
				if err := mongoDbFixtureLoadInviteCodes(); err != nil { mongoDbFixtureHandleFatal(err) }
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

			// handle reload config signal (notify on syscall.SIGHUP)
			// *** not implemented yet ***
		}
	}()

	log.Infof("%s: starting gRPC server at port [%s] ...",metaServiceName,metaServicePort)
	run(metaServicePort); select {}
}

func run(port string) string {

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil { log.Fatal(err) }

	var srv *grpc.Server
	if metaGRPCStatsDisabled := _getDotEnvVariable("DISABLE_STATS"); metaGRPCStatsDisabled == "0" {
		log.Infof("%s: gRPC stats enabled.",metaServiceName)
		srv = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	} else {
		srv = grpc.NewServer()
	}

	userInviteCodeSVC := &UserInviteCodeServiceServer{}
	rfpb.RegisterUserInviteCodeServiceServer(srv, userInviteCodeSVC)
	rfpbh.RegisterHealthServer(srv, userInviteCodeSVC)
	reflection.Register(srv) // activate reflections for grpc

	go srv.Serve(l)

	log.Infof("%s: send SIG.TERM or SIG.INT (CTRL+c) to quit this gRPC endpoint ...",metaServiceName)

	return l.Addr().String()
}

//
// -- gRPC Method Stack 1/n :: HealthCheck(s)
//

func (u UserInviteCodeServiceServer) Check(_ context.Context, _ *rfpbh.HealthCheckRequest) (*rfpbh.HealthCheckResponse, error) {
	return &rfpbh.HealthCheckResponse{Status: rfpbh.HealthCheckResponse_SERVING}, nil
}

func (u UserInviteCodeServiceServer) Watch(_ *rfpbh.HealthCheckRequest, _ rfpbh.Health_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "health check via Watch not implemented")
}

//
// -- gRPC Method Stack 2/n :: MetaOps
//

func (u UserInviteCodeServiceServer) GetVersion(_ context.Context, _ *rfpb.VersionReq) (*rfpb.VersionRes, error) {

	_handleSimLoadLatency()
	return &rfpb.VersionRes{Version: fmt.Sprintf("v%s", metaServiceVersion)}, nil
}

func (u UserInviteCodeServiceServer) CreateInviteCode(_ context.Context, req *rfpb.CreateInviteCodeReq) (*rfpb.CreateInviteCodeRes, error) {

	_handleSimLoadLatency()

	// essentially doing req.GetInviteCode to access the struct with a nil check
	metaCode := req.GetInviteCode()
	metaData := UserInviteCode{
		MetaCode: metaCode.GetMetaCode(),
		MetaForAppRole: metaCode.GetMetaForAppRole(),
		MetaValidFrom: time.Unix(metaCode.GetMetaValidFrom().Seconds, 0),
		MetaValidTo: time.Unix(metaCode.GetMetaValidTo().Seconds, 0),
		CreatedAt: time.Now(),
		IsDeleted: metaCode.GetIsDeleted(),
		IsFixture: metaCode.GetIsFixture(),
		IsTest: metaCode.GetIsTest(),
	}

	log.Infof("%s: CreateInviteCode: receive gRPC invite-guid: %s",metaServiceName,metaData.MetaCode)
	result, err := metaMongoDbCollection.InsertOne(metaMongoDbContext, metaData)
	if err != nil {
		log.Warnf("%s: mongodb: error during gRPC based insert operation of invite-guid xxx",metaServiceName)
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	metaCode.Id = result.InsertedID.(primitive.ObjectID).Hex()
	log.Infof("%s: CreateInviteCode: persist gRPC oid: %s",metaServiceName,metaCode.Id)

	return &rfpb.CreateInviteCodeRes{ InviteCode: metaCode }, nil
}

func (u UserInviteCodeServiceServer) GetInviteCode(_ context.Context, req *rfpb.GetInviteCodeReq) (*rfpb.GetInviteCodeRes, error) {

	_handleSimLoadLatency()

	// convert string id (from proto) to mongoDB ObjectId
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert object-id to document-id",metaServiceName)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	log.Infof("%s: GetInviteCode: receive gRPC invite-oid: %s",metaServiceName,req.GetId())
	result := metaMongoDbCollection.FindOne(metaMongoDbContext, bson.M{"_id": oid, "is_deleted": false})

	metaCode := UserInviteCode{}
	if err := result.Decode(&metaCode); err != nil {
		log.Warnf("%s: mongodb: unable to find document with object-id %s",metaServiceName,req.GetId())
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Get Code Fail (!) -> Error: %v", err))
	}

	// prepare some core type variables
	tsMetaValidFrom, _ := ptypes.TimestampProto(metaCode.MetaValidFrom)
	tsMetaValidTo,   _ := ptypes.TimestampProto(metaCode.MetaValidTo)
	tsMetaCreatedAt, _ := ptypes.TimestampProto(metaCode.CreatedAt)
	// Cast to GetInviteCodeRes type
	response := &rfpb.GetInviteCodeRes{
		InviteCode: &rfpb.UserInviteCode{
			Id:       oid.Hex(),
			MetaCode: metaCode.MetaCode,
			MetaForAppRole: metaCode.MetaForAppRole,
			MetaValidFrom: tsMetaValidFrom,
			MetaValidTo: tsMetaValidTo,
			CreatedAt: tsMetaCreatedAt,
			IsFixture: metaCode.IsFixture,
		},
	}

	return response, nil
}

func (u UserInviteCodeServiceServer) DeleteInviteCode(_ context.Context, req *rfpb.DeleteInviteCodeReq) (*rfpb.DeleteInviteCodeRes, error) {

	_handleSimLoadLatency()

	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert object-id to document-id",metaServiceName)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	result := metaMongoDbCollection.FindOneAndUpdate(metaMongoDbContext,
		bson.M{"_id": oid, "is_deleted": false},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": time.Now(),
		}},  options.FindOneAndUpdate().SetReturnDocument(1))


	decoded := UserInviteCode{}
	err = result.Decode(&decoded)
	if err != nil {
		log.Warnf("%s: mongodb: unable to find invite-code with supplied ID: %s",metaServiceName,oid)
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Delete Code Fail (!) -> Error: %v", err))
	}

	return &rfpb.DeleteInviteCodeRes{ Success: true }, nil
}

func (u UserInviteCodeServiceServer) UpdateInviteCode(_ context.Context, req *rfpb.UpdateInviteCodeReq) (*rfpb.UpdateInviteCodeRes, error) {

	_handleSimLoadLatency()

	metaCode := req.GetInviteCode()
	oid, err := primitive.ObjectIDFromHex(metaCode.GetId())
	if err != nil {
		log.Warnf("%s: mongodb: unable to convert oid (raw=%s)",metaServiceName,metaCode.GetId())
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	res := metaMongoDbCollection.FindOneAndUpdate(metaMongoDbContext,
		bson.M{"_id": oid, "is_deleted": false},
		bson.M{"$set": bson.M{
				"meta_code": metaCode.MetaCode,
				"meta_for_app_role": metaCode.MetaForAppRole,
				"meta_valid_from": metaCode.MetaValidFrom.AsTime(),
				"meta_valid_to": metaCode.MetaValidTo.AsTime(),
				"updated_at": ptypes.TimestampNow(),
			}},  options.FindOneAndUpdate().SetReturnDocument(1))

	decoded := UserInviteCode{}
	err = res.Decode(&decoded)
	if err != nil {
		log.Warnf("%s: mongodb: unable to find document with oid: %s",metaServiceName,oid)
		return nil, status.Errorf(codes.NotFound,fmt.Sprintf("Update Code Fail (!) -> Error: %v", err))
	}

	tsMetaValidFrom, _ := ptypes.TimestampProto(decoded.MetaValidFrom)
	tsMetaValidTo,   _ := ptypes.TimestampProto(decoded.MetaValidTo)

	return &rfpb.UpdateInviteCodeRes{
		InviteCode: &rfpb.UserInviteCode{
			Id:       		decoded.ID.Hex(),
			MetaCode: 		decoded.MetaCode,
			MetaForAppRole: decoded.MetaForAppRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:  	tsMetaValidTo,
		},
	}, nil
}

func (u UserInviteCodeServiceServer) ListInviteCodes(_ *rfpb.ListInviteCodeReq, stream rfpb.UserInviteCodeService_ListInviteCodesServer) error {

	_handleSimLoadLatency()

	data := &UserInviteCode{}
	cursor, err := metaMongoDbCollection.Find(metaMongoDbContext, bson.M{"is_deleted": false})
	if err != nil {
		log.Warnf("%s: mongodb: unable to find invite-code(s)",metaServiceName)
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	};  defer cursor.Close(metaMongoDbContext)

	for cursor.Next(metaMongoDbContext) {
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}
		// prepare some core type variables
		tsMetaValidFrom, _ := ptypes.TimestampProto(data.MetaValidFrom)
		tsMetaValidTo,   _ := ptypes.TimestampProto(data.MetaValidTo)
		// if no error is found send invite codes via stream
		_ = stream.Send(&rfpb.ListInviteCodeRes{
			InviteCode: &rfpb.UserInviteCode{
				Id:             data.ID.Hex(),
				MetaCode:       data.MetaCode,
				MetaForAppRole: data.MetaForAppRole,
				MetaValidFrom:  tsMetaValidFrom,
				MetaValidTo:    tsMetaValidTo,
			},
		})
	}

	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unkown cursor error: %v", err))
	}

	return nil
}

func (u UserInviteCodeServiceServer) ListFilteredInviteCodes(req *rfpb.ListFilteredInviteCodeReq, stream rfpb.UserInviteCodeService_ListFilteredInviteCodesServer) error {

	_handleSimLoadLatency()

	data := &UserInviteCode{}
	cursor, err := metaMongoDbCollection.Find(metaMongoDbContext, _getBSONFilterByRequest(req))
	if err != nil {
		log.Warnf("%s: mongodb: unable to find invite-code(s)",metaServiceName)
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error: %v", err))
	}

	defer cursor.Close(metaMongoDbContext)
	for cursor.Next(metaMongoDbContext) {
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}
		// prepare some core type variables
		tsMetaValidFrom, _ := ptypes.TimestampProto(data.MetaValidFrom)
		tsMetaValidTo,   _ := ptypes.TimestampProto(data.MetaValidTo)
		// if no error is found send filter results (invite codes) via stream
		_ = stream.Send(&rfpb.ListFilteredInviteCodeRes{
			InviteCode: &rfpb.UserInviteCode{
				Id:             data.ID.Hex(),
				MetaCode:       data.MetaCode,
				MetaForAppRole: data.MetaForAppRole,
				MetaValidFrom:  tsMetaValidFrom,
				MetaValidTo:    tsMetaValidTo,
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

func mongoDbFixtureLoadInviteCodes() error {

	var (
		maxAdminCodes = 10
		maxPowerUserCodes = 3
		maxTeacherCodes = 99
		maxViewerCodes = 5
	)

	if err = mongoDbFixtureClean(); err != nil { return err }

	// -- load [user_invitation_codes] metaMongoDbCollection for [admins] --
	log.Infof("%s: mongodb: generate [admin] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	for c := 0; c < maxAdminCodes; c++ {
		if _genUserInviteCode("admin") != nil {
			return err
		}
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [directors (powerUsers)] --
	log.Infof("%s: mongodb: generate [director] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	for c := 0; c < maxPowerUserCodes; c++ {
		if _genUserInviteCode("director") != nil {
			return err
		}
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [teacher] --
	log.Infof("%s: mongodb: generate [teacher] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	for c := 0; c < maxTeacherCodes; c++ {
		if _genUserInviteCode("teacher") != nil {
			return err
		}
	}

	// -- load [user_invitation_codes] metaMongoDbCollection for [viewer] --
	log.Infof("%s: mongodb: generate [viewer] fixtures in metaMongoDbCollection [%s]",metaServiceName,metaMongoDbCollectionTbl)
	for c := 0; c < maxViewerCodes; c++ {
		if _genUserInviteCode("viewer") != nil {
			return err
		}
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
			Keys: bsonx.Doc{{Key: "meta_valid_from", Value: bsonx.Int32(-1)}},
			Options: options.Index().SetName("meta_valid_from_idx").SetSparse(true),
		},
		{
			Keys: bsonx.Doc{{Key: "meta_valid_to", Value: bsonx.Int32(-1)}},
			Options: options.Index().SetName("meta_valid_to_idx").SetSparse(true),
		},
        {
			Keys: bsonx.Doc{{Key: "meta_code", Value: bsonx.String("text")}},
			Options: options.Index().SetName("meta_code_idx").SetUnique(true),
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

func _genUserInviteCodeULID() string {

	t := time.Now().UTC()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)

	return fmt.Sprintf("%s", ulid.MustNew(ulid.Timestamp(t), entropy))
}

func _genUserInviteCode(forRole string) error {

	inviteCode := _genUserInviteCodeULID()
	createdAt  := time.Now()
	validFrom  := createdAt.Add(time.Second * time.Duration(rand.Float32() * 60))
	validTo    := validFrom.Add(time.Hour * 8760) // add 1 year in hours

	_, err = metaMongoDbCollection.InsertOne(metaMongoDbContext, UserInviteCode{
		IsFixture: true,
		MetaCode: inviteCode,
		MetaForAppRole: forRole,
		MetaValidFrom: validFrom,
		MetaValidTo: validTo,
		CreatedAt: createdAt,
	}); if err != nil {
		log.Infof("%s: mongodb: add code [%s]-[%s] failed, code may already exists ...",metaServiceName,inviteCode,forRole)
		return err
	}

	log.Infof("%s: mongodb: add code [%s]-[%s]",metaServiceName,inviteCode,forRole)

	return nil
}

func _getBSONFilterByRequest(req *rfpb.ListFilteredInviteCodeReq) *bson.M {

	dataFilter := &bson.M{}
	metaFilter := req.GetFilter()

	if metaFilter.GetMetaCode() != "" && metaFilter.GetMetaForAppRole() != "" {
		log.Infof("filter [%s] for MetaCode AND [%s] for MetaForAppRole found",metaFilter.GetMetaCode(),metaFilter.GetMetaForAppRole())
		dataFilter = &bson.M{
			"meta_for_app_role" : metaFilter.GetMetaForAppRole(),
			"meta_code"         : metaFilter.GetMetaCode(),
			"is_deleted"        : false }

	} else if metaFilter.GetMetaCode() != "" {
		log.Infof("filter [%s] for MetaCode found",metaFilter.GetMetaCode())
		dataFilter = &bson.M{
			"meta_code"         : metaFilter.GetMetaCode(),
			"is_deleted"        : false }

	} else if metaFilter.GetMetaForAppRole() != "" {
		log.Infof("filter [%s] for MetaForAppRole found",metaFilter.GetMetaForAppRole())
		dataFilter = &bson.M{
			"meta_for_app_role" : metaFilter.GetMetaForAppRole(),
			"is_deleted"        : false }
	}

	return dataFilter
}

//
// -- sidekick stack for env related helper methods
//

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

//
// -- sidekick stack for load-test helper methods
//
func _handleSimLoadLatency() bool {

	if  extraLatency == time.Duration(0) {
		return false
	}

	time.Sleep(extraLatency)

	return true
}