package feishu

import (
	"context"
	"testing"
)

func Test_app_GetAllGroupChats(t *testing.T) {
	fsApp := testNewCustomApp()

	_, err := fsApp.getAppAccessTokenWithContext(context.Background())
	requireNil(t, err)

	fsApp.opt.debug = true

	// groupChatsResp, err := fsApp.GetAllGroupChats()
	// requireNil(t, err)
	//
	// logIndent(t, groupChatsResp)

	// {
	// 	groupChats := make([]GroupChat, 0, 12)
	// 	pageSize := 1
	// 	groupChatsResp, err := fsApp.GetAllGroupChats(WithGetAllGroupChatsPageSize(pageSize))
	// 	requireNil(t, err)
	// 	groupChats = append(groupChats, groupChatsResp.Items...)
	//
	// 	for groupChatsResp.HasMore {
	// 		opts := []GetAllGroupChatsOption{
	// 			WithGetAllGroupChatsPageSize(pageSize),
	// 			WithGetAllGroupChatsPageToken(groupChatsResp.PageToken),
	// 		}
	// 		groupChatsResp, err = fsApp.GetAllGroupChats(opts...)
	// 		requireNil(t, err)
	// 		groupChats = append(groupChats, groupChatsResp.Items...)
	// 	}
	//
	// 	logIndent(t, groupChats)
	// }

	{
		groupChats := make([]GroupChat, 0, 12)
		pageSize := 1
		opts := []GetAllGroupChatsOption{
			WithGetAllGroupChatsPageSize(pageSize),
			WithGetAllGroupChatsOwnerIDType(UnionID),
		}
		groupChatsResp, err := fsApp.GetAllGroupChats(opts...)
		requireNil(t, err)
		groupChats = append(groupChats, groupChatsResp.Items...)

		for groupChatsResp.HasMore {
			opts := []GetAllGroupChatsOption{
				WithGetAllGroupChatsPageSize(pageSize),
				WithGetAllGroupChatsNextPage(groupChatsResp),
			}
			groupChatsResp, err = fsApp.GetAllGroupChats(opts...)
			requireNil(t, err)
			groupChats = append(groupChats, groupChatsResp.Items...)
		}

		logIndent(t, groupChats)
	}
}
