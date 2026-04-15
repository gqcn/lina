// This file defines the unified notify service component and shared transport models.

package notify

import "github.com/gogf/gf/v2/os/gtime"

// Service provides unified notification orchestration and inbox facade operations.
type Service struct{}

// SendInput defines one unified notification send request.
type SendInput struct {
	// ChannelKey is the logical notification channel key.
	ChannelKey string
	// PluginID is the optional source plugin identifier for plugin-originated messages.
	PluginID string
	// SourceType identifies the originating business source type.
	SourceType SourceType
	// SourceID identifies the originating business record.
	SourceID string
	// CategoryCode identifies the message category exposed to inbox consumers.
	CategoryCode CategoryCode
	// Title is the message title displayed to recipients.
	Title string
	// Content is the message body stored in the notify message record.
	Content string
	// Payload carries optional structured message metadata.
	Payload map[string]any
	// SenderUserID is the optional sender user identifier.
	SenderUserID int64
	// RecipientUserIDs is the ordered recipient user identifier list for inbox delivery.
	RecipientUserIDs []int64
}

// SendOutput defines one unified notification send result.
type SendOutput struct {
	// MessageID is the created notify message identifier.
	MessageID int64
	// DeliveryCount is the number of created delivery rows.
	DeliveryCount int
}

// NoticePublishInput defines one notice publication fan-out request.
type NoticePublishInput struct {
	// NoticeID is the published notice identifier.
	NoticeID int64
	// Title is the notice title.
	Title string
	// Content is the notice body content.
	Content string
	// CategoryCode is the inbox category mapped from notice type.
	CategoryCode CategoryCode
	// SenderUserID is the user who created or published the notice.
	SenderUserID int64
}

// InboxListInput defines the inbox list query input.
type InboxListInput struct {
	// UserID is the current inbox user identifier.
	UserID int64
	// PageNum is the 1-based page number.
	PageNum int
	// PageSize is the requested page size.
	PageSize int
}

// InboxListOutput defines the inbox list query result.
type InboxListOutput struct {
	// List is the ordered inbox message slice.
	List []*InboxListItem
	// Total is the total number of matching inbox rows before pagination.
	Total int
}

// InboxListItem defines one inbox list item exposed through the user message facade.
type InboxListItem struct {
	// Id is the notify delivery identifier exposed as the inbox message ID.
	Id int64
	// UserID is the inbox owner user identifier.
	UserID int64
	// Title is the message title displayed in the inbox.
	Title string
	// Type is the legacy message type value: 1=通知 2=公告.
	Type int
	// SourceType is the originating business source type.
	SourceType string
	// SourceID is the legacy numeric source identifier used by current previews.
	SourceID int64
	// IsRead reports whether the inbox row has been marked as read.
	IsRead int
	// ReadAt is the optional read timestamp.
	ReadAt *gtime.Time
	// CreatedAt is the inbox delivery creation timestamp.
	CreatedAt *gtime.Time
}

// New creates and returns a new notify service instance.
func New() *Service {
	return &Service{}
}
