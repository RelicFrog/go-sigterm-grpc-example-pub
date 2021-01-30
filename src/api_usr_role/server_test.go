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
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"io"
	"net"
	"os"
	"regexp"
	"testing"
	"time"
)

const metaTestScope = "api_usr_role_tests"

var ctx = context.Background()

//
// -- core test helper methods :: *.n
//

func tearDBUp() {

	mongoDbInitCon()
}

func tearDBDown(t *testing.T) {

	err = mongoDbFixtureClean()
	if err != nil { t.Fatal(err) }
	mongoDbCloseCon()
}

func grpcDialer() func(context.Context, string) (net.Conn, error) {

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	rfpb.RegisterUserRoleServiceServer(server, &UserRoleServiceServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

//
// -- core test wrapper :: primary call entrypoint for all tests
//

func TestMain(m *testing.M) {

	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {

	if metaMongoDbUsr = _getDotEnvVariable("DB_MONGO_USR"); metaMongoDbUsr == "" {
		fmt.Printf("%s: mongoDB-Service-User not set <exit>",metaTestScope)
		os.Exit(1)
	}

	if metaMongoDbPwd = _getDotEnvVariable("DB_MONGO_PWD"); metaMongoDbPwd == "" {
		fmt.Printf("%s: mongoDB-Password not set <exit>",metaTestScope)
		os.Exit(1)
	}

	if metaMongoDbPDB = _getDotEnvVariable("DB_MONGO_PDB"); metaMongoDbPDB == "" {
		fmt.Printf("%s: mongoDB primary service db not found <exit>",metaTestScope)
		os.Exit(1)
	}

	if metaMongoDbLnk = _getDotEnvVariable("DB_MONGO_LNK"); metaMongoDbLnk == "" {
		fmt.Printf("%s: mongoDB connection link not found <exit>",metaTestScope)
		os.Exit(1)
	}

	if metaServicePort = _getDotEnvVariable("PORT"); metaServicePort == "" {
		fmt.Printf("%s: service port definition missing <exit>",metaTestScope)
		os.Exit(1)
	}

	fmt.Printf("\033[1;36m%s\033[0m", "> Setup completed\n")
}

func teardown() {

	fmt.Printf("\033[1;36m%s\033[0m", "> Teardown completed")
	fmt.Printf("\n")
}

//
// -- core test methods :: UserInviteCodeService GRPC EndPoint(s)
//

func TestRegisterUserInviteCodeServiceServer_GetVersion(t *testing.T) {

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)
	reqGetV := &rfpb.VersionReq{}
	resGetV, err := client.GetVersion(ctx, reqGetV)
	if err != nil { t.Fatalf("GetVersion failed: %v", err) }

	assert.Regexp(t, regexp.MustCompile("v([0-9]+.[0-9]+.[0-9]+)"), resGetV.Version)
}

func TestRegisterUserInviteCodeServiceServer_CreateInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)
	createdAt, _ := ptypes.TimestampProto(time.Now())
	metaRoleName := "ROLE_TEST"
	metaRoleHandle := "role_test"
	metaRoleDescription := fmt.Sprintf("auto-generated test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
	metaRoleIcon := "test_icon"
	metaRoleColorCodeHex := "#000000"
	userRole := rfpb.UserRole{
		IsFixture: true,
		MetaName: metaRoleName,
		MetaDescription: metaRoleDescription,
		MetaAppHandle: metaRoleHandle,
		MetaAppIcon: metaRoleIcon,
		MetaAppColorHex: metaRoleColorCodeHex,
		CreatedAt: createdAt,
	}

	reqCreate := &rfpb.CreateRoleReq{ Role: &userRole }
	resCreate, err := client.CreateRole(ctx, reqCreate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, metaRoleName, resCreate.Role.MetaName)
	assert.Equal(t, metaRoleHandle, resCreate.Role.MetaAppHandle)
	assert.Equal(t, metaRoleDescription, resCreate.Role.MetaDescription)
	assert.Equal(t, metaRoleIcon, resCreate.Role.MetaAppIcon)
	assert.Equal(t, metaRoleColorCodeHex, resCreate.Role.MetaAppColorHex)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_UpdateInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)

	createdAt, _ := ptypes.TimestampProto(time.Now())
	metaRoleName := "ROLE_TEST"
	metaRoleHandle := "role_test"
	metaRoleDescription := fmt.Sprintf("auto-generated test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
	metaRoleIcon := "test_icon"
	metaRoleColorCodeHex := "#000000"
	userRoleOrigin := rfpb.UserRole{
		MetaName: metaRoleName,
		MetaDescription: metaRoleDescription,
		MetaAppHandle: metaRoleHandle,
		MetaAppIcon: metaRoleIcon,
		MetaAppColorHex: metaRoleColorCodeHex,
		CreatedAt: createdAt,
	}

	reqCreate := &rfpb.CreateRoleReq{ Role: &userRoleOrigin }
	resCreate, err := client.CreateRole(ctx, reqCreate)
	if err != nil { tearDBDown(t); t.Fatal(err) }
	assert.Equal(t, metaRoleName, resCreate.Role.MetaName)

	metaNewRoleName := "ROLE_TEST_UPDATED"
	metaNewRoleHandle := "role_test_updated"
	metaNewRoleDescription := fmt.Sprintf("auto-generated new (updated) test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
	metaNewRoleIcon := "test_icon_new"
	metaNewRoleColorCodeHex := "#111111"
	userRoleNew := rfpb.UserRole{
		Id: resCreate.Role.Id,
		MetaName: metaNewRoleName,
		MetaDescription: metaNewRoleDescription,
		MetaAppHandle: metaNewRoleHandle,
		MetaAppIcon: metaNewRoleIcon,
		MetaAppColorHex: metaNewRoleColorCodeHex,
	}

	reqUpdate := &rfpb.UpdateRoleReq{ Role: &userRoleNew }
	resUpdate, err := client.UpdateRole(ctx, reqUpdate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, metaNewRoleName, resUpdate.Role.MetaName)
	assert.Equal(t, metaNewRoleHandle, resUpdate.Role.MetaAppHandle)
	assert.Equal(t, metaNewRoleDescription, resUpdate.Role.MetaDescription)
	assert.Equal(t, metaNewRoleIcon, resUpdate.Role.MetaAppIcon)
	assert.Equal(t, metaNewRoleColorCodeHex, resUpdate.Role.MetaAppColorHex)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_GetInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)

	createdAt, _ := ptypes.TimestampProto(time.Now())
	metaRoleName := "ROLE_TEST"
	metaRoleHandle := "role_test"
	metaRoleDescription := fmt.Sprintf("auto-generated test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
	metaRoleIcon := "test_icon"
	metaRoleColorCodeHex := "#000000"
	userRole := rfpb.UserRole{
		MetaName: metaRoleName,
		MetaDescription: metaRoleDescription,
		MetaAppHandle: metaRoleHandle,
		MetaAppIcon: metaRoleIcon,
		MetaAppColorHex: metaRoleColorCodeHex,
		CreatedAt: createdAt,
	}

	reqCreate := &rfpb.CreateRoleReq{ Role: &userRole }
	resCreate, err := client.CreateRole(ctx, reqCreate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, metaRoleName, resCreate.Role.MetaName)

	reqGetC := &rfpb.GetRoleReq{ Id: resCreate.Role.Id }
	resGetC, err := client.GetRole(ctx, reqGetC)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, resCreate.Role.Id, resGetC.Role.Id)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_DeleteInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)

	createdAt, _ := ptypes.TimestampProto(time.Now())
	metaRoleName := "ROLE_TEST"
	metaRoleHandle := "role_test"
	metaRoleDescription := fmt.Sprintf("auto-generated test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
	metaRoleIcon := "test_icon"
	metaRoleColorCodeHex := "#000000"
	userRole := rfpb.UserRole{
		MetaName: metaRoleName,
		MetaDescription: metaRoleDescription,
		MetaAppHandle: metaRoleHandle,
		MetaAppIcon: metaRoleIcon,
		MetaAppColorHex: metaRoleColorCodeHex,
		CreatedAt: createdAt,
	}

	reqCreate := &rfpb.CreateRoleReq{ Role: &userRole }
	resCreate, err := client.CreateRole(ctx, reqCreate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, metaRoleName, resCreate.Role.MetaName)

	reqDelC := &rfpb.DeleteRoleReq{ Id: resCreate.Role.Id }
	resDelC, err := client.DeleteRole(ctx, reqDelC)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, true, resDelC.Success)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListInviteCodes(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)

	maxRoles := 10
	cntRoles := 0
	for i := 0 ; i < maxRoles ; i++ {
		createdAt, _ := ptypes.TimestampProto(time.Now())
		metaRoleName := fmt.Sprint("ROLE_TEST_%i",i)
		metaRoleHandle := fmt.Sprint("role_test_%i",i)
		metaRoleDescription := fmt.Sprintf("auto-generated test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
		metaRoleIcon := "test_icon"
		metaRoleColorCodeHex := fmt.Sprint("#00000%i",i)
		userRole := rfpb.UserRole{
			MetaName: metaRoleName,
			MetaDescription: metaRoleDescription,
			MetaAppHandle: metaRoleHandle,
			MetaAppIcon: metaRoleIcon,
			MetaAppColorHex: metaRoleColorCodeHex,
			CreatedAt: createdAt,
		}

		reqCreate := &rfpb.CreateRoleReq{ Role: &userRole }
		resCreate, err := client.CreateRole(ctx, reqCreate)
		if err != nil { tearDBDown(t); t.Fatal(err) }

		assert.Equal(t, metaRoleName, resCreate.Role.MetaName)
	}

	reqGetList := &rfpb.ListRoleReq{}
	resGetList, err := client.ListRoles(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	for {

		res, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		assert.True(t, res.Role.MetaName != "")
		assert.True(t, res.Role.MetaDescription != "")
		assert.True(t, res.Role.MetaAppHandle != "")
		assert.True(t, res.Role.MetaAppIcon != "")
		assert.True(t, res.Role.MetaAppColorHex != "")

		cntRoles++
	}

	assert.Equal(t, cntRoles, maxRoles)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListInviteCodesNilOnDeleted(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserRoleServiceClient(conn)

	maxRoles := 5
	cntRoles := 0
	for i := 0 ; i < maxRoles ; i++ {
		createdAt, _ := ptypes.TimestampProto(time.Now())
		metaRoleName := fmt.Sprint("ROLE_TEST_INVISIBLE_%i",i)
		metaRoleHandle := fmt.Sprint("role_test_invisible_%i",i)
		metaRoleDescription := fmt.Sprintf("auto-generated (invisible) test-role [%s] using handle [%s]",metaRoleName,metaRoleHandle);
		metaRoleIcon := "test_icon_invisible"
		metaRoleColorCodeHex := fmt.Sprint("#10000%i",i)
		userRole := rfpb.UserRole{
			IsDeleted: true,
			MetaName: metaRoleName,
			MetaDescription: metaRoleDescription,
			MetaAppHandle: metaRoleHandle,
			MetaAppIcon: metaRoleIcon,
			MetaAppColorHex: metaRoleColorCodeHex,
			CreatedAt: createdAt,
		}

		reqCreate := &rfpb.CreateRoleReq{ Role: &userRole }
		resCreate, err := client.CreateRole(ctx, reqCreate)
		if err != nil { tearDBDown(t); t.Fatal(err) }

		assert.Equal(t, metaRoleName, resCreate.Role.MetaName)
	}

	reqGetList := &rfpb.ListRoleReq{}
	resGetList, err := client.ListRoles(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	for {

		_, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		cntRoles++
	}

	assert.Equal(t, cntRoles, 0)

	tearDBDown(t)
}

//
// -- test helper methods for repeating test content using different params
// -* none *-
//
