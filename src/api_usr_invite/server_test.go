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
	rfpb "api_usr_invite/server/proto"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/oklog/ulid/v2"
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

const metaTestScope = "api_usr_invite_tests"

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
	rfpb.RegisterUserInviteCodeServiceServer(server, &UserInviteCodeServiceServer{})
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

	client := rfpb.NewUserInviteCodeServiceClient(conn)
	reqGetV := &rfpb.VersionReq{}
	resGetV, err := client.GetVersion(ctx, reqGetV)
	if err != nil { t.Fatalf("GetVersion failed: %v", err) }

	assert.Regexp(t, regexp.MustCompile("v([0-9]+.[0-9]+.[0-9]+)"), resGetV.Version)
}

func TestRegisterUserInviteCodeServiceServer_CreateInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	MetaInviteCode := _genUserInviteCodeULID()
	MetaInviteRole := "ninja"
	tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
	tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
	inviteCode := rfpb.UserInviteCode{
		MetaCode:       MetaInviteCode,
		MetaForAppRole: MetaInviteRole,
		MetaValidFrom:  tsMetaValidFrom,
		MetaValidTo:    tsMetaValidTo,
	}

	reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCode }
	resCreate, err := client.CreateInviteCode(ctx, reqCreate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	assert.Equal(t, MetaInviteRole, resCreate.InviteCode.MetaForAppRole)
	assert.Equal(t, tsMetaValidFrom.Seconds, resCreate.InviteCode.MetaValidFrom.Seconds)
	assert.Equal(t, tsMetaValidTo.Seconds, resCreate.InviteCode.MetaValidTo.Seconds)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_UpdateInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	MetaInviteCode := _genUserInviteCodeULID()
	MetaInviteRole := "ninja"
	tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
	tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
	inviteCodeOrigin := rfpb.UserInviteCode{
		MetaCode:       MetaInviteCode,
		MetaForAppRole: MetaInviteRole,
		MetaValidFrom:  tsMetaValidFrom,
		MetaValidTo:    tsMetaValidTo,
	}

	reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
	resCreate, err := client.CreateInviteCode(ctx, reqCreate)
	if err != nil { t.Fatal(err) }
	assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)

	tsMetaNewValidFrom, _ := ptypes.TimestampProto(tsMetaValidFrom.AsTime().Add(time.Hour * time.Duration(1)))
	tsMetaNewValidTo,   _ := ptypes.TimestampProto(tsMetaValidTo.AsTime().Add(time.Hour * time.Duration(2)))

	MetaNewInviteCode := _genUserInviteCodeULID()
	MetaNewInviteRole := "gopher"
	inviteCodeNew := rfpb.UserInviteCode{
		Id:             resCreate.InviteCode.Id,
		MetaCode:       MetaNewInviteCode,
		MetaForAppRole: MetaNewInviteRole,
		MetaValidFrom:  tsMetaNewValidFrom,
		MetaValidTo:    tsMetaNewValidTo,
	}

	reqUpdate := &rfpb.UpdateInviteCodeReq{ InviteCode: &inviteCodeNew }
	resUpdate, err := client.UpdateInviteCode(ctx, reqUpdate)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, MetaNewInviteCode, resUpdate.InviteCode.MetaCode)
	assert.Equal(t, MetaNewInviteRole, resUpdate.InviteCode.MetaForAppRole)
	assert.Equal(t, tsMetaNewValidFrom.Seconds, resUpdate.InviteCode.MetaValidFrom.Seconds)
	assert.Equal(t, tsMetaNewValidTo.Seconds, resUpdate.InviteCode.MetaValidTo.Seconds)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_GetInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	MetaInviteCode := _genUserInviteCodeULID()
	MetaInviteRole := "ninja"
	tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
	tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
	inviteCodeOrigin := rfpb.UserInviteCode{
		MetaCode:       MetaInviteCode,
		MetaForAppRole: MetaInviteRole,
		MetaValidFrom:  tsMetaValidFrom,
		MetaValidTo:    tsMetaValidTo,
	}

	reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
	resCreate, err := client.CreateInviteCode(ctx, reqCreate)
	if err != nil { t.Fatal(err) }
	assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)

	reqGetC := &rfpb.GetInviteCodeReq{ Id: resCreate.InviteCode.Id }
	resGetC, err := client.GetInviteCode(ctx, reqGetC)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, resCreate.InviteCode.Id, resGetC.InviteCode.Id)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_DeleteInviteCode(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	MetaInviteCode := _genUserInviteCodeULID()
	MetaInviteRole := "bad-ninja"
	tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
	tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
	inviteCodeOrigin := rfpb.UserInviteCode{
		MetaCode:       MetaInviteCode,
		MetaForAppRole: MetaInviteRole,
		MetaValidFrom:  tsMetaValidFrom,
		MetaValidTo:    tsMetaValidTo,
	}

	reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
	resCreate, err := client.CreateInviteCode(ctx, reqCreate)
	if err != nil { t.Fatal(err) }
	assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)

	reqDelC := &rfpb.DeleteInviteCodeReq{ Id: resCreate.InviteCode.Id }
	resDelC, err := client.DeleteInviteCode(ctx, reqDelC)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	assert.Equal(t, true, resDelC.Success)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListInviteCodes(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	maxInviteCodes := 10
	cntInviteCodes := 0
	for i := 0 ; i < maxInviteCodes ; i++ {
		MetaInviteCode := _genUserInviteCodeULID()
		MetaInviteRole := "twin-ninja"
		tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
		inviteCodeOrigin := rfpb.UserInviteCode{
			MetaCode:       MetaInviteCode,
			MetaForAppRole: MetaInviteRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:    tsMetaValidTo,
		}

		reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
		resCreate, err := client.CreateInviteCode(ctx, reqCreate)
		if err != nil { t.Fatal(err) }
		assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	}

	reqGetList := &rfpb.ListInviteCodeReq{}
	resGetList, err := client.ListInviteCodes(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	for {

		res, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		_, err = ulid.ParseStrict(res.GetInviteCode().MetaCode)
		assert.True(t, err == nil)

		cntInviteCodes++
	}

	assert.Equal(t, cntInviteCodes, maxInviteCodes)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListInviteCodesNilOnDeleted(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	maxInviteCodes := 3
	cntInviteCodes := 0
	for i := 0 ; i < maxInviteCodes ; i++ {
		MetaInviteCode := _genUserInviteCodeULID()
		MetaInviteRole := "deleted"
		tsMetaDeletedAt, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
		inviteCodeOrigin := rfpb.UserInviteCode{
			MetaCode:       MetaInviteCode,
			MetaForAppRole: MetaInviteRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:    tsMetaValidTo,
			DeletedAt:      tsMetaDeletedAt,
			IsDeleted:      true,
		}

		reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
		resCreate, err := client.CreateInviteCode(ctx, reqCreate)
		if err != nil { t.Fatal(err) }
		assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	}

	reqGetList := &rfpb.ListInviteCodeReq{}
	resGetList, err := client.ListInviteCodes(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	for {

		_, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		cntInviteCodes++
	}

	assert.Equal(t, cntInviteCodes, 0)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListFilteredInviteCodesAdmin(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	inviteCodesRole := "admin"
	inviteCodesMax  := 7

	for i := 0 ; i < inviteCodesMax ; i++ {
		MetaInviteCode := _genUserInviteCodeULID()
		MetaInviteRole := inviteCodesRole
		tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
		inviteCodeOrigin := rfpb.UserInviteCode{
			MetaCode:       MetaInviteCode,
			MetaForAppRole: MetaInviteRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:    tsMetaValidTo,
		}

		reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
		resCreate, err := client.CreateInviteCode(ctx, reqCreate)
		if err != nil { t.Fatal(err) }
		assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	}

	_testFilterForRole(t, conn, inviteCodesRole, inviteCodesMax)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListFilteredInviteCodesViewer(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	inviteCodesRole := "viewer"
	inviteCodesMax  := 5

	for i := 0 ; i < inviteCodesMax ; i++ {
		MetaInviteCode := _genUserInviteCodeULID()
		MetaInviteRole := inviteCodesRole
		tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
		inviteCodeOrigin := rfpb.UserInviteCode{
			MetaCode:       MetaInviteCode,
			MetaForAppRole: MetaInviteRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:    tsMetaValidTo,
		}

		reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
		resCreate, err := client.CreateInviteCode(ctx, reqCreate)
		if err != nil { t.Fatal(err) }
		assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	}

	_testFilterForRole(t, conn, inviteCodesRole, inviteCodesMax)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListFilteredInviteCodesNilOnDeleted(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	inviteCodesRole := "deleted"
	inviteCodesMax  := 3

	for i := 0 ; i < inviteCodesMax ; i++ {
		MetaInviteCode := _genUserInviteCodeULID()
		MetaInviteRole := inviteCodesRole
		tsMetaDeletedAt, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
		tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
		inviteCodeOrigin := rfpb.UserInviteCode{
			MetaCode:       MetaInviteCode,
			MetaForAppRole: MetaInviteRole,
			MetaValidFrom:  tsMetaValidFrom,
			MetaValidTo:    tsMetaValidTo,
			DeletedAt:      tsMetaDeletedAt,
			IsDeleted:      true,
		}

		reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
		resCreate, err := client.CreateInviteCode(ctx, reqCreate)
		if err != nil { t.Fatal(err) }
		assert.Equal(t, MetaInviteCode, resCreate.InviteCode.MetaCode)
	}

	_testFilterForRole(t, conn, inviteCodesRole, 0)

	tearDBDown(t)
}

func TestRegisterUserInviteCodeServiceServer_ListFilteredInviteCodesByGuid(t *testing.T) {

	tearDBUp()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(grpcDialer()))
	if err != nil { log.Fatal(err) }; defer conn.Close()

	client := rfpb.NewUserInviteCodeServiceClient(conn)

	inviteCodeGuid := _genUserInviteCodeULID()
	tsMetaValidFrom, _ := ptypes.TimestampProto(time.Now())
	tsMetaValidTo, _ := ptypes.TimestampProto(time.Now())
	inviteCodeOrigin := rfpb.UserInviteCode{
		MetaCode:       inviteCodeGuid,
		MetaForAppRole: "ninja",
		MetaValidFrom:  tsMetaValidFrom,
		MetaValidTo:    tsMetaValidTo,
	}

	reqCreate := &rfpb.CreateInviteCodeReq{ InviteCode: &inviteCodeOrigin }
	resCreate, err := client.CreateInviteCode(ctx, reqCreate)
	if err != nil { t.Fatal(err) }
	assert.Equal(t, inviteCodeGuid, resCreate.InviteCode.MetaCode)

	_testFilterForCode(t, conn, inviteCodeGuid, 1)

	tearDBDown(t)
}

//
// -- test helper methods for repeating test content using different params
//

func _testFilterForRole(t *testing.T, conn *grpc.ClientConn, inviteCodeRole string, inviteCodeCount int) {

	filter := &rfpb.UserInviteCodeFilter{ MetaForAppRole: inviteCodeRole }
	client := rfpb.NewUserInviteCodeServiceClient(conn)

	reqGetList := &rfpb.ListFilteredInviteCodeReq{ Filter: filter }
	resGetList, err := client.ListFilteredInviteCodes(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	cntInviteCodes := 0

	for {

		res, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		_, err = ulid.ParseStrict(res.GetInviteCode().MetaCode)
		assert.True(t, err == nil)

		cntInviteCodes++
	}

	assert.Equal(t, cntInviteCodes, inviteCodeCount)
}

func _testFilterForCode(t *testing.T, conn *grpc.ClientConn, inviteCode string, inviteCodeCount int) {

	filter := &rfpb.UserInviteCodeFilter{ MetaCode: inviteCode }
	client := rfpb.NewUserInviteCodeServiceClient(conn)

	reqGetList := &rfpb.ListFilteredInviteCodeReq{ Filter: filter }
	resGetList, err := client.ListFilteredInviteCodes(ctx, reqGetList)
	if err != nil { tearDBDown(t); t.Fatal(err) }

	cntInviteCodes := 0

	for {

		res, err := resGetList.Recv()
		if err == io.EOF { break }
		if err != nil { tearDBDown(t); t.Fatal(err) }

		_, err = ulid.ParseStrict(res.GetInviteCode().MetaCode)
		assert.True(t, err == nil)
		cntInviteCodes++
	}

	assert.Equal(t, cntInviteCodes, inviteCodeCount)
}
