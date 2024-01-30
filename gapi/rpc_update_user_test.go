package gapi

import (
	"context"
	"database/sql"
	"testing"
	"time"

	mockdb "github.com/jtaylor-io/safe-as-houses/db/mock"
	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/jtaylor-io/safe-as-houses/pb"
	"github.com/jtaylor-io/safe-as-houses/token"
	"github.com/jtaylor-io/safe-as-houses/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateUserAPI(t *testing.T) {
	user, _ := randomUser(t)

	newName := util.RandomOwner()
	newEmail := util.RandomEmail()
	invalidEmail := "invalid-email"

	testCases := []struct {
		name          string
		req           *pb.UpdateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		buildContext  func(t *testing.T, tokenMaker token.Maker) context.Context
		checkResponse func(t *testing.T, res *pb.UpdateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},

			buildStubs: func(store *mockdb.MockStore) {
				arg := db.UpdateUserParams{
					FullName: sql.NullString{
						String: newName,
						Valid:  true,
					},
					Email: sql.NullString{
						String: newEmail,
						Valid:  true,
					},
					Username: user.Username,
				}

				updatedUser := db.User{
					Username:          user.Username,
					HashedPassword:    user.HashedPassword,
					FullName:          newName,
					Email:             newEmail,
					PasswordChangedAt: user.PasswordChangedAt,
					CreatedAt:         user.CreatedAt,
					IsEmailVerified:   user.IsEmailVerified,
				}
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(updatedUser, nil)
			},
			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(
					t,
					tokenMaker,
					user.Username,
					user.Role,
					time.Minute,
				)
			},

			checkResponse: func(t *testing.T, res *pb.UpdateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				updatedUser := res.GetUser()
				require.Equal(t, user.Username, updatedUser.Username)
				require.Equal(t, newName, updatedUser.FullName)
				require.Equal(t, newEmail, updatedUser.Email)
			},
		},
		{
			name: "UserNotFound",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},

			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(
					t,
					tokenMaker,
					user.Username,
					user.Role,
					time.Minute,
				)
			},

			checkResponse: func(t *testing.T, _ *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.NotFound, st.Code())
			},
		},
		{
			name: "ExpiredToken",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},

			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(
					t,
					tokenMaker,
					user.Username,
					user.Role,
					-time.Minute,
				)
			},

			checkResponse: func(t *testing.T, _ *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "NoAuthorization",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &newEmail,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},

			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return context.Background()
			},

			checkResponse: func(t *testing.T, _ *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Unauthenticated, st.Code())
			},
		},
		{
			name: "InvalidEmail",
			req: &pb.UpdateUserRequest{
				Username: user.Username,
				FullName: &newName,
				Email:    &invalidEmail,
			},

			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},

			buildContext: func(t *testing.T, tokenMaker token.Maker) context.Context {
				return newContextWithBearerToken(
					t,
					tokenMaker,
					user.Username,
					user.Role,
					time.Minute,
				)
			},

			checkResponse: func(t *testing.T, _ *pb.UpdateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			},
		},
		// {
		// 	name: "Internal",
		// 	req: &pb.CreateUserRequest{
		// 		Username: user.Username,
		// 		FullName: user.FullName,
		// 		Email:    user.Email,
		// 		Password: password,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
		// 		store.EXPECT().
		// 			CreateUserTx(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.CreateUserTxResult{}, sql.ErrConnDone)
		// 		taskDistributor.EXPECT().
		// 			DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, _ *pb.CreateUserResponse, err error) {
		// 		require.Error(t, err)
		// 		st, ok := status.FromError(err)
		// 		require.True(t, ok)
		// 		require.Equal(t, codes.Internal, st.Code())
		// 	},
		// },
		// {
		// 	name: "InvalidUsername",
		// 	body: gin.H{
		// 		"username":  "non-alphanum->!",
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidPassword",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password[:9],
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidFullName",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": "",
		// 		"email":     user.Email,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InvalidEmail",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     strings.ReplaceAll("@", user.Email, ""),
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(0)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusBadRequest, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "InternalError",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.User{}, sql.ErrConnDone)
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusInternalServerError, recorder.Code)
		// 	},
		// },
		// {
		// 	name: "DuplicateUsername",
		// 	body: gin.H{
		// 		"username":  user.Username,
		// 		"password":  password,
		// 		"full_name": user.FullName,
		// 		"email":     user.Email,
		// 	},
		//
		// 	buildStubs: func(store *mockdb.MockStore) {
		// 		store.EXPECT().
		// 			CreateUser(gomock.Any(), gomock.Any()).
		// 			Times(1).
		// 			Return(db.User{}, &pq.Error{Code: "23505"})
		// 	},
		//
		// 	checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
		// 		require.Equal(t, http.StatusForbidden, recorder.Code)
		// 	},
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storeCtrl := gomock.NewController(t)
			defer storeCtrl.Finish()
			store := mockdb.NewMockStore(storeCtrl)

			tc.buildStubs(store)

			server := newTestServer(t, store, nil)
			ctx := tc.buildContext(t, server.tokenMaker)
			res, err := server.UpdateUser(ctx, tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
