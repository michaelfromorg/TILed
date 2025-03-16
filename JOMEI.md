# jomei/notionapi

Documentation for `jomei/notionapi`.

---

Repository
github.com/jomei/notionapi
Links

    Open Source Insights Logo Open Source Insights

README ¶
notionapi

GitHub tag (latest SemVer) Go Reference Test

This is a Golang implementation of an API client for the Notion API.
Supported APIs

It supports all APIs of the Notion API version 2022-06-28.
Installation

go get github.com/jomei/notionapi

Usage

First, please follow the Getting Started Guide to obtain an integration token.
Initialization

Import this library and initialize the API client using the obtained integration token.

import "github.com/jomei/notionapi"

client := notionapi.NewClient("your_integration_token")

Calling the API

You can use the methods of the initialized client to call the Notion API. Here is an example of how to retrieve a page:

page, err := client.Page.Get(context.Background(), "your_page_id")
if err != nil {
    // Handle the error
}

Documentation ¶
Index ¶

    type AndCompoundFilter
        func (f AndCompoundFilter) MarshalJSON() ([]byte, error)
    type Annotations
    type AppendBlockChildrenRequest
    type AppendBlockChildrenResponse
        func (r *AppendBlockChildrenResponse) UnmarshalJSON(data []byte) error
    type Audio
        func (i Audio) GetURL() string
    type AudioBlock
        func (b AudioBlock) GetRichTextString() string
    type AuthenticationClient
        func (cc *AuthenticationClient) CreateToken(ctx context.Context, request *TokenCreateRequest) (*TokenCreateResponse, error)
    type AuthenticationService
    type BasicBlock
        func (b BasicBlock) GetArchived() bool
        func (b BasicBlock) GetCreatedBy() *User
        func (b BasicBlock) GetCreatedTime() *time.Time
        func (b BasicBlock) GetHasChildren() bool
        func (b BasicBlock) GetID() BlockID
        func (b BasicBlock) GetLastEditedBy() *User
        func (b BasicBlock) GetLastEditedTime() *time.Time
        func (b BasicBlock) GetObject() ObjectType
        func (b BasicBlock) GetParent() *Parent
        func (b BasicBlock) GetRichTextString() string
        func (b BasicBlock) GetType() BlockType
    type Block
    type BlockClient
        func (bc *BlockClient) AppendChildren(ctx context.Context, id BlockID, requestBody *AppendBlockChildrenRequest) (*AppendBlockChildrenResponse, error)
        func (bc *BlockClient) Delete(ctx context.Context, id BlockID) (Block, error)
        func (bc *BlockClient) Get(ctx context.Context, id BlockID) (Block, error)
        func (bc *BlockClient) GetChildren(ctx context.Context, id BlockID, pagination *Pagination) (*GetChildrenResponse, error)
        func (bc *BlockClient) Update(ctx context.Context, id BlockID, requestBody *BlockUpdateRequest) (Block, error)
    type BlockFile
    type BlockID
        func (bID BlockID) String() string
    type BlockService
    type BlockType
        func (bt BlockType) String() string
    type BlockUpdateRequest
    type Blocks
        func (b *Blocks) UnmarshalJSON(data []byte) error
    type Bookmark
    type BookmarkBlock
        func (b BookmarkBlock) GetRichTextString() string
    type Bot
    type Breadcrumb
    type BreadcrumbBlock
    type BulletedListItemBlock
        func (b BulletedListItemBlock) GetRichTextString() string
    type Button
    type ButtonProperty
        func (p ButtonProperty) GetID() string
        func (p ButtonProperty) GetType() PropertyType
    type ButtonPropertyConfig
        func (p ButtonPropertyConfig) GetID() PropertyID
        func (p ButtonPropertyConfig) GetType() PropertyConfigType
    type Callout
    type CalloutBlock
        func (c CalloutBlock) GetRichTextString() string
    type CheckboxFilterCondition
    type CheckboxProperty
        func (p CheckboxProperty) GetID() string
        func (p CheckboxProperty) GetType() PropertyType
    type CheckboxPropertyConfig
        func (p CheckboxPropertyConfig) GetID() PropertyID
        func (p CheckboxPropertyConfig) GetType() PropertyConfigType
    type ChildDatabaseBlock
    type ChildPageBlock
    type Client
        func NewClient(token Token, opts ...ClientOption) *Client
    type ClientOption
        func WithHTTPClient(client *http.Client) ClientOption
        func WithOAuthAppCredentials(id, secret string) ClientOption
        func WithRetry(retries int) ClientOption
        func WithVersion(version string) ClientOption
    type Code
    type CodeBlock
    type Color
        func (c Color) MarshalText() ([]byte, error)
        func (c Color) String() string
    type Column
    type ColumnBlock
    type ColumnList
    type ColumnListBlock
    type Comment
    type CommentClient
        func (cc *CommentClient) Create(ctx context.Context, requestBody *CommentCreateRequest) (*Comment, error)
        func (cc *CommentClient) Get(ctx context.Context, id BlockID, pagination *Pagination) (*CommentQueryResponse, error)
    type CommentCreateRequest
    type CommentID
        func (cID CommentID) String() string
    type CommentQueryResponse
    type CommentService
    type CompoundFilter
    type Condition
    type CreatedByProperty
        func (p CreatedByProperty) GetID() string
        func (p CreatedByProperty) GetType() PropertyType
    type CreatedByPropertyConfig
        func (p CreatedByPropertyConfig) GetID() PropertyID
        func (p CreatedByPropertyConfig) GetType() PropertyConfigType
    type CreatedTimeProperty
        func (p CreatedTimeProperty) GetID() string
        func (p CreatedTimeProperty) GetType() PropertyType
    type CreatedTimePropertyConfig
        func (p CreatedTimePropertyConfig) GetID() PropertyID
        func (p CreatedTimePropertyConfig) GetType() PropertyConfigType
    type Cursor
        func (c Cursor) String() string
    type Database
        func (db *Database) GetObject() ObjectType
    type DatabaseClient
        func (dc *DatabaseClient) Create(ctx context.Context, requestBody *DatabaseCreateRequest) (*Database, error)
        func (dc *DatabaseClient) Get(ctx context.Context, id DatabaseID) (*Database, error)
        func (dc *DatabaseClient) Query(ctx context.Context, id DatabaseID, requestBody *DatabaseQueryRequest) (*DatabaseQueryResponse, error)
        func (dc *DatabaseClient) Update(ctx context.Context, id DatabaseID, requestBody *DatabaseUpdateRequest) (*Database, error)
    type DatabaseCreateRequest
    type DatabaseID
        func (dID DatabaseID) String() string
    type DatabaseMention
    type DatabaseQueryRequest
        func (qr *DatabaseQueryRequest) MarshalJSON() ([]byte, error)
    type DatabaseQueryResponse
    type DatabaseService
    type DatabaseUpdateRequest
    type Date
        func (d Date) MarshalText() ([]byte, error)
        func (d *Date) String() string
        func (d *Date) UnmarshalText(data []byte) error
    type DateFilterCondition
    type DateObject
    type DateProperty
        func (p DateProperty) GetID() string
        func (p DateProperty) GetType() PropertyType
    type DatePropertyConfig
        func (p DatePropertyConfig) GetID() PropertyID
        func (p DatePropertyConfig) GetType() PropertyConfigType
    type DiscussionID
        func (dID DiscussionID) String() string
    type Divider
    type DividerBlock
    type DownloadableFileBlock
    type DualProperty
    type EmailProperty
        func (p EmailProperty) GetID() string
        func (p EmailProperty) GetType() PropertyType
    type EmailPropertyConfig
        func (p EmailPropertyConfig) GetID() PropertyID
        func (p EmailPropertyConfig) GetType() PropertyConfigType
    type Embed
    type EmbedBlock
        func (b EmbedBlock) GetRichTextString() string
    type Emoji
    type Equation
    type EquationBlock
        func (b EquationBlock) GetRichTextString() string
    type Error
        func (e *Error) Error() string
    type ErrorCode
    type ExternalAccount
    type File
    type FileBlock
        func (b *FileBlock) GetExpiryTime() *time.Time
        func (b FileBlock) GetRichTextString() string
        func (b *FileBlock) GetURL() string
    type FileObject
    type FileType
    type FilesFilterCondition
    type FilesProperty
        func (p FilesProperty) GetID() string
        func (p FilesProperty) GetType() PropertyType
    type FilesPropertyConfig
        func (p FilesPropertyConfig) GetID() PropertyID
        func (p FilesPropertyConfig) GetType() PropertyConfigType
    type Filter
    type FilterOperator
    type FormatType
        func (ft FormatType) String() string
    type Formula
    type FormulaConfig
    type FormulaFilterCondition
    type FormulaProperty
        func (p FormulaProperty) GetID() string
        func (p FormulaProperty) GetType() PropertyType
    type FormulaPropertyConfig
        func (p FormulaPropertyConfig) GetID() PropertyID
        func (p FormulaPropertyConfig) GetType() PropertyConfigType
    type FormulaType
    type FunctionType
        func (ft FunctionType) String() string
    type GetChildrenResponse
    type GroupConfig
    type Heading
    type Heading1Block
        func (h Heading1Block) GetRichTextString() string
    type Heading2Block
        func (h Heading2Block) GetRichTextString() string
    type Heading3Block
        func (h Heading3Block) GetRichTextString() string
    type Icon
        func (i Icon) GetURL() string
    type Image
        func (i Image) GetURL() string
    type ImageBlock
        func (b *ImageBlock) GetExpiryTime() *time.Time
        func (b ImageBlock) GetRichTextString() string
        func (b *ImageBlock) GetURL() string
    type LastEditedByProperty
        func (p LastEditedByProperty) GetID() string
        func (p LastEditedByProperty) GetType() PropertyType
    type LastEditedByPropertyConfig
        func (p LastEditedByPropertyConfig) GetID() PropertyID
        func (p LastEditedByPropertyConfig) GetType() PropertyConfigType
    type LastEditedTimeProperty
        func (p LastEditedTimeProperty) GetID() string
        func (p LastEditedTimeProperty) GetType() PropertyType
    type LastEditedTimePropertyConfig
        func (p LastEditedTimePropertyConfig) GetID() PropertyID
        func (p LastEditedTimePropertyConfig) GetType() PropertyConfigType
    type Link
    type LinkPreview
    type LinkPreviewBlock
        func (b LinkPreviewBlock) GetRichTextString() string
    type LinkToPage
    type LinkToPageBlock
    type ListItem
    type Mention
    type MentionType
        func (mType MentionType) String() string
    type MultiSelectFilterCondition
    type MultiSelectProperty
        func (p MultiSelectProperty) GetID() string
        func (p MultiSelectProperty) GetType() PropertyType
    type MultiSelectPropertyConfig
        func (p MultiSelectPropertyConfig) GetID() PropertyID
        func (p MultiSelectPropertyConfig) GetType() PropertyConfigType
    type NumberFilterCondition
    type NumberFormat
    type NumberProperty
        func (p NumberProperty) GetID() string
        func (p NumberProperty) GetType() PropertyType
    type NumberPropertyConfig
        func (p NumberPropertyConfig) GetID() PropertyID
        func (p NumberPropertyConfig) GetType() PropertyConfigType
    type NumberedListItemBlock
        func (n NumberedListItemBlock) GetRichTextString() string
    type Object
    type ObjectID
        func (oID ObjectID) String() string
    type ObjectType
        func (ot ObjectType) String() string
    type Option
    type OrCompoundFilter
        func (f OrCompoundFilter) MarshalJSON() ([]byte, error)
    type Owner
    type Page
        func (p *Page) GetObject() ObjectType
    type PageClient
        func (pc *PageClient) Create(ctx context.Context, requestBody *PageCreateRequest) (*Page, error)
        func (pc *PageClient) Get(ctx context.Context, id PageID) (*Page, error)
        func (pc *PageClient) Update(ctx context.Context, id PageID, request *PageUpdateRequest) (*Page, error)
    type PageCreateRequest
    type PageID
        func (pID PageID) String() string
    type PageMention
    type PageService
    type PageUpdateRequest
    type Pagination
        func (p *Pagination) ToQuery() map[string]string
    type Paragraph
    type ParagraphBlock
        func (p ParagraphBlock) GetRichTextString() string
    type Parent
    type ParentType
    type Pdf
    type PdfBlock
        func (b *PdfBlock) GetExpiryTime() *time.Time
        func (b PdfBlock) GetRichTextString() string
        func (b *PdfBlock) GetURL() string
    type PeopleFilterCondition
    type PeopleProperty
        func (p PeopleProperty) GetID() string
        func (p PeopleProperty) GetType() PropertyType
    type PeoplePropertyConfig
        func (p PeoplePropertyConfig) GetID() PropertyID
        func (p PeoplePropertyConfig) GetType() PropertyConfigType
    type Person
    type PhoneNumberProperty
        func (p PhoneNumberProperty) GetID() string
        func (p PhoneNumberProperty) GetType() PropertyType
    type PhoneNumberPropertyConfig
        func (p PhoneNumberPropertyConfig) GetID() PropertyID
        func (p PhoneNumberPropertyConfig) GetType() PropertyConfigType
    type Properties
        func (p *Properties) UnmarshalJSON(data []byte) error
    type Property
    type PropertyArray
        func (arr *PropertyArray) UnmarshalJSON(data []byte) error
    type PropertyConfig
    type PropertyConfigType
    type PropertyConfigs
        func (p *PropertyConfigs) UnmarshalJSON(data []byte) error
    type PropertyFilter
    type PropertyID
        func (pID PropertyID) String() string
    type PropertyType
    type Quote
    type QuoteBlock
        func (q QuoteBlock) GetRichTextString() string
    type RateLimitedError
        func (e *RateLimitedError) Error() string
    type Relation
    type RelationConfig
    type RelationConfigType
        func (rp RelationConfigType) String() string
    type RelationFilterCondition
    type RelationObject
    type RelationProperty
        func (p RelationProperty) GetID() string
        func (p RelationProperty) GetType() PropertyType
    type RelationPropertyConfig
        func (p RelationPropertyConfig) GetID() PropertyID
        func (p RelationPropertyConfig) GetType() PropertyConfigType
    type RichText
    type RichTextProperty
        func (p RichTextProperty) GetID() string
        func (p RichTextProperty) GetType() PropertyType
    type RichTextPropertyConfig
        func (p RichTextPropertyConfig) GetID() PropertyID
        func (p RichTextPropertyConfig) GetType() PropertyConfigType
    type Rollup
    type RollupConfig
    type RollupFilterCondition
    type RollupProperty
        func (p RollupProperty) GetID() string
        func (p RollupProperty) GetType() PropertyType
    type RollupPropertyConfig
        func (p RollupPropertyConfig) GetID() PropertyID
        func (p RollupPropertyConfig) GetType() PropertyConfigType
    type RollupSubfilterCondition
    type RollupType
    type SearchClient
        func (sc *SearchClient) Do(ctx context.Context, request *SearchRequest) (*SearchResponse, error)
    type SearchFilter
    type SearchRequest
    type SearchResponse
        func (sr *SearchResponse) UnmarshalJSON(data []byte) error
    type SearchService
    type Select
    type SelectFilterCondition
    type SelectProperty
        func (p SelectProperty) GetID() string
        func (p SelectProperty) GetType() PropertyType
    type SelectPropertyConfig
        func (p SelectPropertyConfig) GetID() PropertyID
        func (p SelectPropertyConfig) GetType() PropertyConfigType
    type SingleProperty
    type SortObject
    type SortOrder
    type Status
    type StatusConfig
    type StatusFilterCondition
    type StatusProperty
        func (p StatusProperty) GetID() string
        func (p StatusProperty) GetType() PropertyType
    type StatusPropertyConfig
        func (p StatusPropertyConfig) GetID() PropertyID
        func (p StatusPropertyConfig) GetType() PropertyConfigType
    type Synced
    type SyncedBlock
    type SyncedFrom
    type Table
    type TableBlock
    type TableOfContents
    type TableOfContentsBlock
    type TableRow
    type TableRowBlock
    type Template
    type TemplateBlock
        func (b TemplateBlock) GetRichTextString() string
    type TemplateMention
    type TemplateMentionType
        func (tMType TemplateMentionType) String() string
    type Text
    type TextFilterCondition
    type TextProperty
        func (p TextProperty) GetID() string
        func (p TextProperty) GetType() PropertyType
    type TimestampFilter
    type TimestampType
    type TitleProperty
        func (p TitleProperty) GetID() string
        func (p TitleProperty) GetType() PropertyType
    type TitlePropertyConfig
        func (p TitlePropertyConfig) GetID() PropertyID
        func (p TitlePropertyConfig) GetType() PropertyConfigType
    type ToDo
    type ToDoBlock
        func (t ToDoBlock) GetRichTextString() string
    type Toggle
    type ToggleBlock
        func (b ToggleBlock) GetRichTextString() string
    type Token
        func (it Token) String() string
    type TokenCreateError
        func (e *TokenCreateError) Error() string
    type TokenCreateRequest
    type TokenCreateResponse
    type URLProperty
        func (p URLProperty) GetID() string
        func (p URLProperty) GetType() PropertyType
    type URLPropertyConfig
        func (p URLPropertyConfig) GetID() PropertyID
        func (p URLPropertyConfig) GetType() PropertyConfigType
    type UniqueID
        func (uID UniqueID) String() string
    type UniqueIDConfig
    type UniqueIDProperty
        func (p UniqueIDProperty) GetID() string
        func (p UniqueIDProperty) GetType() PropertyType
    type UniqueIDPropertyConfig
        func (p UniqueIDPropertyConfig) GetID() PropertyID
        func (p UniqueIDPropertyConfig) GetType() PropertyConfigType
    type UniqueIdFilterCondition
    type UnsupportedBlock
    type User
    type UserClient
        func (uc *UserClient) Get(ctx context.Context, id UserID) (*User, error)
        func (uc *UserClient) List(ctx context.Context, pagination *Pagination) (*UsersListResponse, error)
        func (uc *UserClient) Me(ctx context.Context) (*User, error)
    type UserID
        func (uID UserID) String() string
    type UserService
    type UserType
    type UsersListResponse
    type Verification
    type VerificationProperty
        func (p VerificationProperty) GetID() string
        func (p VerificationProperty) GetType() PropertyType
    type VerificationPropertyConfig
        func (p VerificationPropertyConfig) GetID() PropertyID
        func (p VerificationPropertyConfig) GetType() PropertyConfigType
    type VerificationState
        func (vs VerificationState) String() string
    type Video
    type VideoBlock
        func (b VideoBlock) GetRichTextString() string

Constants ¶

This section is empty.
Variables ¶

This section is empty.
Functions ¶

This section is empty.
Types ¶
type AndCompoundFilter ¶ added in v1.8.5

type AndCompoundFilter []Filter

func (AndCompoundFilter) MarshalJSON ¶ added in v1.8.5

func (f AndCompoundFilter) MarshalJSON() ([]byte, error)

type Annotations ¶

type Annotations struct {
	Bold          bool  `json:"bold"`
	Italic        bool  `json:"italic"`
	Strikethrough bool  `json:"strikethrough"`
	Underline     bool  `json:"underline"`
	Code          bool  `json:"code"`
	Color         Color `json:"color,omitempty"`
}

type AppendBlockChildrenRequest ¶

type AppendBlockChildrenRequest struct {
	// Append new children after a specific block. If empty, new children with be appended to the bottom of the parent block.
	After BlockID `json:"after,omitempty"`
	// Child content to append to a container block as an array of block objects.
	Children []Block `json:"children"`
}

type AppendBlockChildrenResponse ¶ added in v1.6.0

type AppendBlockChildrenResponse struct {
	Object  ObjectType `json:"object"`
	Results []Block    `json:"results"`
}

func (*AppendBlockChildrenResponse) UnmarshalJSON ¶ added in v1.6.0

func (r *AppendBlockChildrenResponse) UnmarshalJSON(data []byte) error

type Audio ¶ added in v1.10.3

type Audio struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

func (Audio) GetURL ¶ added in v1.10.3

func (i Audio) GetURL() string

GetURL returns the external or internal URL depending on the image type.
type AudioBlock ¶ added in v1.10.3

type AudioBlock struct {
	BasicBlock
	Audio Audio `json:"audio"`
}

func (AudioBlock) GetRichTextString ¶ added in v1.12.10

func (b AudioBlock) GetRichTextString() string

type AuthenticationClient ¶ added in v1.12.0

type AuthenticationClient struct {
	// contains filtered or unexported fields
}

func (*AuthenticationClient) CreateToken ¶ added in v1.12.0

func (cc *AuthenticationClient) CreateToken(ctx context.Context, request *TokenCreateRequest) (*TokenCreateResponse, error)

Creates an access token that a third-party service can use to authenticate with Notion.

See https://developers.notion.com/reference/create-a-token
type AuthenticationService ¶ added in v1.12.0

type AuthenticationService interface {
	CreateToken(ctx context.Context, request *TokenCreateRequest) (*TokenCreateResponse, error)
}

type BasicBlock ¶ added in v1.7.0

type BasicBlock struct {
	Object         ObjectType `json:"object"`
	ID             BlockID    `json:"id,omitempty"`
	Type           BlockType  `json:"type"`
	CreatedTime    *time.Time `json:"created_time,omitempty"`
	LastEditedTime *time.Time `json:"last_edited_time,omitempty"`
	CreatedBy      *User      `json:"created_by,omitempty"`
	LastEditedBy   *User      `json:"last_edited_by,omitempty"`
	HasChildren    bool       `json:"has_children,omitempty"`
	Archived       bool       `json:"archived,omitempty"`
	Parent         *Parent    `json:"parent,omitempty"`
}

BasicBlock defines the common fields of all Notion block types. See https://developers.notion.com/reference/block for the list. BasicBlock implements the Block interface.
func (BasicBlock) GetArchived ¶ added in v1.7.1

func (b BasicBlock) GetArchived() bool

func (BasicBlock) GetCreatedBy ¶ added in v1.8.1

func (b BasicBlock) GetCreatedBy() *User

func (BasicBlock) GetCreatedTime ¶ added in v1.7.1

func (b BasicBlock) GetCreatedTime() *time.Time

func (BasicBlock) GetHasChildren ¶ added in v1.7.1

func (b BasicBlock) GetHasChildren() bool

func (BasicBlock) GetID ¶ added in v1.7.1

func (b BasicBlock) GetID() BlockID

func (BasicBlock) GetLastEditedBy ¶ added in v1.8.1

func (b BasicBlock) GetLastEditedBy() *User

func (BasicBlock) GetLastEditedTime ¶ added in v1.7.1

func (b BasicBlock) GetLastEditedTime() *time.Time

func (BasicBlock) GetObject ¶ added in v1.7.1

func (b BasicBlock) GetObject() ObjectType

func (BasicBlock) GetParent ¶ added in v1.12.8

func (b BasicBlock) GetParent() *Parent

func (BasicBlock) GetRichTextString ¶ added in v1.12.10

func (b BasicBlock) GetRichTextString() string

func (BasicBlock) GetType ¶ added in v1.7.0

func (b BasicBlock) GetType() BlockType

type Block ¶

type Block interface {
	GetType() BlockType
	GetID() BlockID
	GetObject() ObjectType
	GetCreatedTime() *time.Time
	GetLastEditedTime() *time.Time
	GetCreatedBy() *User
	GetLastEditedBy() *User
	GetHasChildren() bool
	GetArchived() bool
	GetParent() *Parent
	GetRichTextString() string
}

type BlockClient ¶

type BlockClient struct {
	// contains filtered or unexported fields
}

func (*BlockClient) AppendChildren ¶

func (bc *BlockClient) AppendChildren(ctx context.Context, id BlockID, requestBody *AppendBlockChildrenRequest) (*AppendBlockChildrenResponse, error)

For blocks that allow children, we allow up to two levels of nesting in a single request.

See https://developers.notion.com/reference/patch-block-children
func (*BlockClient) Delete ¶ added in v1.7.0

func (bc *BlockClient) Delete(ctx context.Context, id BlockID) (Block, error)

Sets a Block object, including page blocks, to archived: true using the ID specified. Note: in the Notion UI application, this moves the block to the "Trash" where it can still be accessed and restored.

To restore the block with the API, use the Update a block or Update page respectively.

See https://developers.notion.com/reference/delete-a-block
func (*BlockClient) Get ¶ added in v1.4.0

func (bc *BlockClient) Get(ctx context.Context, id BlockID) (Block, error)

Retrieves a Block object using the ID specified.

Get https://developers.notion.com/reference/retrieve-a-block
func (*BlockClient) GetChildren ¶

func (bc *BlockClient) GetChildren(ctx context.Context, id BlockID, pagination *Pagination) (*GetChildrenResponse, error)

Returns a paginated array of child block objects contained in the block using the ID specified. In order to receive a complete representation of a block, you may need to recursively retrieve the block children of child blocks.

See https://developers.notion.com/reference/get-block-children
func (*BlockClient) Update ¶ added in v1.4.0

func (bc *BlockClient) Update(ctx context.Context, id BlockID, requestBody *BlockUpdateRequest) (Block, error)

Updates the content for the specified block_id based on the block type. Supported fields based on the block object type (see Block object for available fields and the expected input for each field).

Note: The update replaces the entire value for a given field. If a field is omitted (ex: omitting checked when updating a to_do block), the value will not be changed.

See https://developers.notion.com/reference/update-a-block
type BlockFile ¶ added in v1.5.0

type BlockFile struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type BlockID ¶

type BlockID string

func (BlockID) String ¶

func (bID BlockID) String() string

type BlockService ¶

type BlockService interface {
	AppendChildren(context.Context, BlockID, *AppendBlockChildrenRequest) (*AppendBlockChildrenResponse, error)
	Get(context.Context, BlockID) (Block, error)
	GetChildren(context.Context, BlockID, *Pagination) (*GetChildrenResponse, error)
	Update(ctx context.Context, id BlockID, request *BlockUpdateRequest) (Block, error)
	Delete(context.Context, BlockID) (Block, error)
}

type BlockType ¶

type BlockType string

const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeHeading1  BlockType = "heading_1"
	BlockTypeHeading2  BlockType = "heading_2"
	BlockTypeHeading3  BlockType = "heading_3"

	BlockTypeBulletedListItem BlockType = "bulleted_list_item"
	BlockTypeNumberedListItem BlockType = "numbered_list_item"

	BlockTypeToDo          BlockType = "to_do"
	BlockTypeToggle        BlockType = "toggle"
	BlockTypeChildPage     BlockType = "child_page"
	BlockTypeChildDatabase BlockType = "child_database"

	BlockTypeEmbed           BlockType = "embed"
	BlockTypeImage           BlockType = "image"
	BlockTypeVideo           BlockType = "video"
	BlockTypeFile            BlockType = "file"
	BlockTypePdf             BlockType = "pdf"
	BlockTypeBookmark        BlockType = "bookmark"
	BlockTypeCode            BlockType = "code"
	BlockTypeDivider         BlockType = "divider"
	BlockTypeCallout         BlockType = "callout"
	BlockTypeQuote           BlockType = "quote"
	BlockTypeTableOfContents BlockType = "table_of_contents"
	BlockTypeEquation        BlockType = "equation"
	BlockTypeBreadcrumb      BlockType = "breadcrumb"
	BlockTypeColumn          BlockType = "column"
	BlockTypeColumnList      BlockType = "column_list"
	BlockTypeLinkPreview     BlockType = "link_preview"
	BlockTypeLinkToPage      BlockType = "link_to_page"
	BlockTypeTemplate        BlockType = "template"
	BlockTypeSyncedBlock     BlockType = "synced_block"
	BlockTypeTableBlock      BlockType = "table"
	BlockTypeTableRowBlock   BlockType = "table_row"
	BlockTypeUnsupported     BlockType = "unsupported"
)

See https://developers.notion.com/reference/block
func (BlockType) String ¶

func (bt BlockType) String() string

type BlockUpdateRequest ¶ added in v1.4.0

type BlockUpdateRequest struct {
	Paragraph        *Paragraph `json:"paragraph,omitempty"`
	Heading1         *Heading   `json:"heading_1,omitempty"`
	Heading2         *Heading   `json:"heading_2,omitempty"`
	Heading3         *Heading   `json:"heading_3,omitempty"`
	BulletedListItem *ListItem  `json:"bulleted_list_item,omitempty"`
	NumberedListItem *ListItem  `json:"numbered_list_item,omitempty"`
	Code             *Code      `json:"code,omitempty"`
	ToDo             *ToDo      `json:"to_do,omitempty"`
	Toggle           *Toggle    `json:"toggle,omitempty"`
	Embed            *Embed     `json:"embed,omitempty"`
	Image            *Image     `json:"image,omitempty"`
	Video            *Video     `json:"video,omitempty"`
	File             *BlockFile `json:"file,omitempty"`
	Pdf              *Pdf       `json:"pdf,omitempty"`
	Bookmark         *Bookmark  `json:"bookmark,omitempty"`
	Template         *Template  `json:"template,omitempty"`
	Callout          *Callout   `json:"callout,omitempty"`
	Equation         *Equation  `json:"equation,omitempty"`
	Quote            *Quote     `json:"quote,omitempty"`
	TableRow         *TableRow  `json:"table_row,omitempty"`
}

type Blocks ¶ added in v1.7.1

type Blocks []Block

func (*Blocks) UnmarshalJSON ¶ added in v1.7.1

func (b *Blocks) UnmarshalJSON(data []byte) error

type Bookmark ¶ added in v1.5.0

type Bookmark struct {
	Caption []RichText `json:"caption,omitempty"`
	URL     string     `json:"url"`
}

type BookmarkBlock ¶ added in v1.5.0

type BookmarkBlock struct {
	BasicBlock
	Bookmark Bookmark `json:"bookmark"`
}

func (BookmarkBlock) GetRichTextString ¶ added in v1.12.10

func (b BookmarkBlock) GetRichTextString() string

type Bot ¶

type Bot struct {
	Owner         Owner  `json:"owner"`
	WorkspaceName string `json:"workspace_name"`
}

type Breadcrumb ¶ added in v1.7.0

type Breadcrumb struct {
}

type BreadcrumbBlock ¶ added in v1.7.0

type BreadcrumbBlock struct {
	BasicBlock
	Breadcrumb Breadcrumb `json:"breadcrumb"`
}

type BulletedListItemBlock ¶

type BulletedListItemBlock struct {
	BasicBlock
	BulletedListItem ListItem `json:"bulleted_list_item"`
}

func (BulletedListItemBlock) GetRichTextString ¶ added in v1.12.10

func (b BulletedListItemBlock) GetRichTextString() string

type Button ¶ added in v1.13.0

type Button struct {
}

type ButtonProperty ¶ added in v1.13.0

type ButtonProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Button Button       `json:"button"`
}

func (ButtonProperty) GetID ¶ added in v1.13.0

func (p ButtonProperty) GetID() string

func (ButtonProperty) GetType ¶ added in v1.13.0

func (p ButtonProperty) GetType() PropertyType

type ButtonPropertyConfig ¶ added in v1.13.2

type ButtonPropertyConfig struct {
	ID     PropertyID         `json:"id"`
	Type   PropertyConfigType `json:"type"`
	Button struct{}           `json:"button"`
}

func (ButtonPropertyConfig) GetID ¶ added in v1.13.2

func (p ButtonPropertyConfig) GetID() PropertyID

func (ButtonPropertyConfig) GetType ¶ added in v1.13.2

func (p ButtonPropertyConfig) GetType() PropertyConfigType

type Callout ¶ added in v1.5.3

type Callout struct {
	RichText []RichText `json:"rich_text"`
	Icon     *Icon      `json:"icon,omitempty"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type CalloutBlock ¶ added in v1.5.3

type CalloutBlock struct {
	BasicBlock
	Callout Callout `json:"callout"`
}

func (CalloutBlock) GetRichTextString ¶ added in v1.12.10

func (c CalloutBlock) GetRichTextString() string

type CheckboxFilterCondition ¶ added in v1.2.0

type CheckboxFilterCondition struct {
	Equals       bool `json:"equals,omitempty"`
	DoesNotEqual bool `json:"does_not_equal,omitempty"`
}

type CheckboxProperty ¶

type CheckboxProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	Checkbox bool         `json:"checkbox"`
}

func (CheckboxProperty) GetID ¶ added in v1.12.9

func (p CheckboxProperty) GetID() string

func (CheckboxProperty) GetType ¶

func (p CheckboxProperty) GetType() PropertyType

type CheckboxPropertyConfig ¶ added in v1.2.0

type CheckboxPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	Checkbox struct{}           `json:"checkbox"`
}

func (CheckboxPropertyConfig) GetID ¶ added in v1.13.0

func (p CheckboxPropertyConfig) GetID() PropertyID

func (CheckboxPropertyConfig) GetType ¶ added in v1.2.0

func (p CheckboxPropertyConfig) GetType() PropertyConfigType

type ChildDatabaseBlock ¶ added in v1.5.1

type ChildDatabaseBlock struct {
	BasicBlock
	ChildDatabase struct {
		Title string `json:"title"`
	} `json:"child_database"`
}

type ChildPageBlock ¶

type ChildPageBlock struct {
	BasicBlock
	ChildPage struct {
		Title string `json:"title"`
	} `json:"child_page"`
}

type Client ¶

type Client struct {
	Token Token

	Database       DatabaseService
	Block          BlockService
	Page           PageService
	User           UserService
	Search         SearchService
	Comment        CommentService
	Authentication AuthenticationService
	// contains filtered or unexported fields
}

func NewClient ¶

func NewClient(token Token, opts ...ClientOption) *Client

type ClientOption ¶

type ClientOption func(*Client)

ClientOption to configure API client
func WithHTTPClient ¶

func WithHTTPClient(client *http.Client) ClientOption

WithHTTPClient overrides the default http.Client
func WithOAuthAppCredentials ¶ added in v1.12.3

func WithOAuthAppCredentials(id, secret string) ClientOption

WithOAuthAppCredentials sets the OAuth app ID and secret to use when fetching a token from Notion.
func WithRetry ¶ added in v1.10.0

func WithRetry(retries int) ClientOption

WithRetry overrides the default number of max retry attempts on 429 errors
func WithVersion ¶

func WithVersion(version string) ClientOption

WithVersion overrides the Notion API version
type Code ¶ added in v1.5.1

type Code struct {
	RichText []RichText `json:"rich_text"`
	Caption  []RichText `json:"caption,omitempty"`
	Language string     `json:"language"`
}

type CodeBlock ¶ added in v1.5.1

type CodeBlock struct {
	BasicBlock
	Code Code `json:"code"`
}

type Color ¶

type Color string

const (
	ColorDefault           Color = "default"
	ColorGray              Color = "gray"
	ColorBrown             Color = "brown"
	ColorOrange            Color = "orange"
	ColorYellow            Color = "yellow"
	ColorGreen             Color = "green"
	ColorBlue              Color = "blue"
	ColorPurple            Color = "purple"
	ColorPink              Color = "pink"
	ColorRed               Color = "red"
	ColorDefaultBackground Color = "default_background"
	ColorGrayBackground    Color = "gray_background"
	ColorBrownBackground   Color = "brown_background"
	ColorOrangeBackground  Color = "orange_background"
	ColorYellowBackground  Color = "yellow_background"
	ColorGreenBackground   Color = "green_background"
	ColorBlueBackground    Color = "blue_background"
	ColorPurpleBackground  Color = "purple_background"
	ColorPinkBackground    Color = "pink_background"
	ColorRedBackground     Color = "red_background"
)

func (Color) MarshalText ¶ added in v1.4.0

func (c Color) MarshalText() ([]byte, error)

func (Color) String ¶

func (c Color) String() string

type Column ¶ added in v1.7.0

type Column struct {
	// Children should at least have 1 block when appending.
	Children Blocks `json:"children"`
}

type ColumnBlock ¶ added in v1.7.0

type ColumnBlock struct {
	BasicBlock
	Column Column `json:"column"`
}

type ColumnList ¶ added in v1.7.0

type ColumnList struct {
	// Children can only contain column blocks
	// Children should have at least 2 blocks when appending.
	Children Blocks `json:"children"`
}

type ColumnListBlock ¶ added in v1.7.0

type ColumnListBlock struct {
	BasicBlock
	ColumnList ColumnList `json:"column_list"`
}

type Comment ¶ added in v1.9.2

type Comment struct {
	Object         ObjectType   `json:"object"`
	ID             ObjectID     `json:"id"`
	DiscussionID   DiscussionID `json:"discussion_id"`
	CreatedTime    time.Time    `json:"created_time"`
	LastEditedTime time.Time    `json:"last_edited_time"`
	CreatedBy      User         `json:"created_by,omitempty"`
	RichText       []RichText   `json:"rich_text"`
	Parent         Parent       `json:"parent"`
}

type CommentClient ¶ added in v1.9.2

type CommentClient struct {
	// contains filtered or unexported fields
}

func (*CommentClient) Create ¶ added in v1.9.2

func (cc *CommentClient) Create(ctx context.Context, requestBody *CommentCreateRequest) (*Comment, error)

Creates a comment in a page or existing discussion thread.

There are two locations you can add a new comment to: 1. A page 2. An existing discussion thread

If the intention is to add a new comment to a page, a parent object must be provided in the body params. Alternatively, if a new comment is being added to an existing discussion thread, the discussion_id string must be provided in the body params. Exactly one of these parameters must be provided.

See https://developers.notion.com/reference/create-a-comment
func (*CommentClient) Get ¶ added in v1.9.2

func (cc *CommentClient) Get(ctx context.Context, id BlockID, pagination *Pagination) (*CommentQueryResponse, error)

Retrieves a list of un-resolved Comment objects from a page or block.

See https://developers.notion.com/reference/retrieve-a-comment
type CommentCreateRequest ¶ added in v1.9.2

type CommentCreateRequest struct {
	Parent       Parent       `json:"parent,omitempty"`
	DiscussionID DiscussionID `json:"discussion_id,omitempty"`
	RichText     []RichText   `json:"rich_text"`
}

CommentCreateRequest represents the request body for CommentClient.Create.
type CommentID ¶ added in v1.9.2

type CommentID string

func (CommentID) String ¶ added in v1.9.2

func (cID CommentID) String() string

type CommentQueryResponse ¶ added in v1.9.2

type CommentQueryResponse struct {
	Object     ObjectType `json:"object"`
	Results    []Comment  `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

type CommentService ¶ added in v1.9.2

type CommentService interface {
	Create(ctx context.Context, request *CommentCreateRequest) (*Comment, error)
	Get(context.Context, BlockID, *Pagination) (*CommentQueryResponse, error)
}

type CompoundFilter ¶

type CompoundFilter map[FilterOperator][]PropertyFilter

type Condition ¶

type Condition string

const (
	ConditionEquals         Condition = "equals"
	ConditionDoesNotEqual   Condition = "does_not_equal"
	ConditionContains       Condition = "contains"
	ConditionDoesNotContain Condition = "does_not_contain"
	ConditionDoesStartsWith Condition = "starts_with"
	ConditionDoesEndsWith   Condition = "ends_with"
	ConditionDoesIsEmpty    Condition = "is_empty"
	ConditionGreaterThan    Condition = "greater_than"
	ConditionLessThan       Condition = "less_than"

	ConditionGreaterThanOrEqualTo Condition = "greater_than_or_equal_to"
	ConditionLessThanOrEqualTo    Condition = "greater_than_or_equal_to"

	ConditionBefore     Condition = "before"
	ConditionAfter      Condition = "after"
	ConditionOnOrBefore Condition = "on_or_before"
	ConditionOnOrAfter  Condition = "on_or_after"
	ConditionPastWeek   Condition = "past_week"
	ConditionPastMonth  Condition = "past_month"
	ConditionPastYear   Condition = "past_year"
	ConditionNextWeek   Condition = "next_week"
	ConditionNextMonth  Condition = "next_month"
	ConditionNextYear   Condition = "next_year"

	ConditionText     Condition = "text"
	ConditionCheckbox Condition = "checkbox"
	ConditionNumber   Condition = "number"
	ConditionDate     Condition = "date"
)

type CreatedByProperty ¶

type CreatedByProperty struct {
	ID        ObjectID     `json:"id,omitempty"`
	Type      PropertyType `json:"type,omitempty"`
	CreatedBy User         `json:"created_by"`
}

func (CreatedByProperty) GetID ¶ added in v1.12.9

func (p CreatedByProperty) GetID() string

func (CreatedByProperty) GetType ¶

func (p CreatedByProperty) GetType() PropertyType

type CreatedByPropertyConfig ¶ added in v1.2.0

type CreatedByPropertyConfig struct {
	ID        PropertyID         `json:"id"`
	Type      PropertyConfigType `json:"type"`
	CreatedBy struct{}           `json:"created_by"`
}

func (CreatedByPropertyConfig) GetID ¶ added in v1.13.0

func (p CreatedByPropertyConfig) GetID() PropertyID

func (CreatedByPropertyConfig) GetType ¶ added in v1.2.0

func (p CreatedByPropertyConfig) GetType() PropertyConfigType

type CreatedTimeProperty ¶

type CreatedTimeProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	CreatedTime time.Time    `json:"created_time"`
}

func (CreatedTimeProperty) GetID ¶ added in v1.12.9

func (p CreatedTimeProperty) GetID() string

func (CreatedTimeProperty) GetType ¶

func (p CreatedTimeProperty) GetType() PropertyType

type CreatedTimePropertyConfig ¶ added in v1.2.0

type CreatedTimePropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	CreatedTime struct{}           `json:"created_time"`
}

func (CreatedTimePropertyConfig) GetID ¶ added in v1.13.0

func (p CreatedTimePropertyConfig) GetID() PropertyID

func (CreatedTimePropertyConfig) GetType ¶ added in v1.2.0

func (p CreatedTimePropertyConfig) GetType() PropertyConfigType

type Cursor ¶

type Cursor string

func (Cursor) String ¶

func (c Cursor) String() string

type Database ¶

type Database struct {
	Object         ObjectType `json:"object"`
	ID             ObjectID   `json:"id"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	CreatedBy      User       `json:"created_by,omitempty"`
	LastEditedBy   User       `json:"last_edited_by,omitempty"`
	Title          []RichText `json:"title"`
	Parent         Parent     `json:"parent"`
	URL            string     `json:"url"`
	PublicURL      string     `json:"public_url"`
	// Properties is a map of property configurations that defines what Page.Properties each page of the database can use
	Properties  PropertyConfigs `json:"properties"`
	Description []RichText      `json:"description"`
	IsInline    bool            `json:"is_inline"`
	Archived    bool            `json:"archived"`
	Icon        *Icon           `json:"icon,omitempty"`
	Cover       *Image          `json:"cover,omitempty"`
}

func (*Database) GetObject ¶

func (db *Database) GetObject() ObjectType

type DatabaseClient ¶

type DatabaseClient struct {
	// contains filtered or unexported fields
}

func (*DatabaseClient) Create ¶ added in v1.4.0

func (dc *DatabaseClient) Create(ctx context.Context, requestBody *DatabaseCreateRequest) (*Database, error)

Creates a database as a subpage in the specified parent page, with the specified properties schema. Currently, the parent of a new database must be a Notion page.

See https://developers.notion.com/reference/create-a-database
func (*DatabaseClient) Get ¶

func (dc *DatabaseClient) Get(ctx context.Context, id DatabaseID) (*Database, error)

See https://developers.notion.com/reference/get-database
func (*DatabaseClient) Query ¶

func (dc *DatabaseClient) Query(ctx context.Context, id DatabaseID, requestBody *DatabaseQueryRequest) (*DatabaseQueryResponse, error)

Gets a list of Pages contained in the database, filtered and ordered according to the filter conditions and sort criteria provided in the request. The response may contain fewer than page_size of results. If the response includes a next_cursor value, refer to the pagination reference for details about how to use a cursor to iterate through the list.

Filters are similar to the filters provided in the Notion UI where the set of filters and filter groups chained by "And" in the UI is equivalent to having each filter in the array of the compound "and" filter. Similar a set of filters chained by "Or" in the UI would be represented as filters in the array of the "or" compound filter.

Filters operate on database properties and can be combined. If no filter is provided, all the pages in the database will be returned with pagination.

See https://developers.notion.com/reference/post-database-query
func (*DatabaseClient) Update ¶ added in v1.3.0

func (dc *DatabaseClient) Update(ctx context.Context, id DatabaseID, requestBody *DatabaseUpdateRequest) (*Database, error)

Update https://developers.notion.com/reference/update-a-database
type DatabaseCreateRequest ¶ added in v1.4.0

type DatabaseCreateRequest struct {
	// A page parent.
	Parent Parent `json:"parent"`
	// Title of database as it appears in Notion. An array of rich text objects.
	Title []RichText `json:"title"`
	// Property schema of database. The keys are the names of properties as they
	// appear in Notion and the values are property schema objects.
	Properties PropertyConfigs `json:"properties"`
	IsInline   bool            `json:"is_inline"`
}

DatabaseCreateRequest represents the request body for DatabaseClient.Create.
type DatabaseID ¶

type DatabaseID string

func (DatabaseID) String ¶

func (dID DatabaseID) String() string

type DatabaseMention ¶ added in v1.9.1

type DatabaseMention struct {
	ID ObjectID `json:"id"`
}

type DatabaseQueryRequest ¶

type DatabaseQueryRequest struct {
	// When supplied, limits which pages are returned based on the filter
	// conditions.
	Filter Filter
	// When supplied, orders the results based on the provided sort criteria.
	Sorts []SortObject `json:"sorts,omitempty"`
	// When supplied, returns a page of results starting after the cursor provided.
	// If not supplied, this endpoint will return the first page of results.
	StartCursor Cursor `json:"start_cursor,omitempty"`
	// The number of items from the full list desired in the response. Maximum: 100
	PageSize int `json:"page_size,omitempty"`
}

DatabaseQueryRequest represents the request body for DatabaseClient.Query.
func (*DatabaseQueryRequest) MarshalJSON ¶ added in v1.2.0

func (qr *DatabaseQueryRequest) MarshalJSON() ([]byte, error)

type DatabaseQueryResponse ¶

type DatabaseQueryResponse struct {
	Object     ObjectType `json:"object"`
	Results    []Page     `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

type DatabaseService ¶

type DatabaseService interface {
	Create(ctx context.Context, request *DatabaseCreateRequest) (*Database, error)
	Query(context.Context, DatabaseID, *DatabaseQueryRequest) (*DatabaseQueryResponse, error)
	Get(context.Context, DatabaseID) (*Database, error)
	Update(context.Context, DatabaseID, *DatabaseUpdateRequest) (*Database, error)
}

type DatabaseUpdateRequest ¶ added in v1.3.0

type DatabaseUpdateRequest struct {
	// An array of rich text objects that represents the title of the database
	// that is displayed in the Notion UI. If omitted, then the database title
	// remains unchanged.
	Title []RichText `json:"title,omitempty"`
	// The properties of a database to be changed in the request, in the form of
	// a JSON object. If updating an existing property, then the keys are the
	// names or IDs of the properties as they appear in Notion, and the values are
	// property schema objects. If adding a new property, then the key is the name
	// of the new database property and the value is a property schema object.
	Properties PropertyConfigs `json:"properties,omitempty"`
}

DatabaseUpdateRequest represents the request body for DatabaseClient.Update.
type Date ¶ added in v1.2.0

type Date time.Time

func (Date) MarshalText ¶ added in v1.2.0

func (d Date) MarshalText() ([]byte, error)

func (*Date) String ¶ added in v1.2.0

func (d *Date) String() string

func (*Date) UnmarshalText ¶ added in v1.3.1

func (d *Date) UnmarshalText(data []byte) error

type DateFilterCondition ¶ added in v1.2.0

type DateFilterCondition struct {
	Equals     *Date     `json:"equals,omitempty"`
	Before     *Date     `json:"before,omitempty"`
	After      *Date     `json:"after,omitempty"`
	OnOrBefore *Date     `json:"on_or_before,omitempty"`
	OnOrAfter  *Date     `json:"on_or_after,omitempty"`
	PastWeek   *struct{} `json:"past_week,omitempty"`
	PastMonth  *struct{} `json:"past_month,omitempty"`
	PastYear   *struct{} `json:"past_year,omitempty"`
	NextWeek   *struct{} `json:"next_week,omitempty"`
	NextMonth  *struct{} `json:"next_month,omitempty"`
	NextYear   *struct{} `json:"next_year,omitempty"`
	IsEmpty    bool      `json:"is_empty,omitempty"`
	IsNotEmpty bool      `json:"is_not_empty,omitempty"`
}

type DateObject ¶ added in v1.2.0

type DateObject struct {
	Start *Date `json:"start"`
	End   *Date `json:"end"`
}

type DateProperty ¶

type DateProperty struct {
	ID   ObjectID     `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	Date *DateObject  `json:"date"`
}

func (DateProperty) GetID ¶ added in v1.12.9

func (p DateProperty) GetID() string

func (DateProperty) GetType ¶

func (p DateProperty) GetType() PropertyType

type DatePropertyConfig ¶ added in v1.2.0

type DatePropertyConfig struct {
	ID   PropertyID         `json:"id,omitempty"`
	Type PropertyConfigType `json:"type"`
	Date struct{}           `json:"date"`
}

func (DatePropertyConfig) GetID ¶ added in v1.13.0

func (p DatePropertyConfig) GetID() PropertyID

func (DatePropertyConfig) GetType ¶ added in v1.2.0

func (p DatePropertyConfig) GetType() PropertyConfigType

type DiscussionID ¶ added in v1.9.2

type DiscussionID string

func (DiscussionID) String ¶ added in v1.9.2

func (dID DiscussionID) String() string

type Divider ¶ added in v1.5.3

type Divider struct {
}

type DividerBlock ¶ added in v1.5.3

type DividerBlock struct {
	BasicBlock
	Divider Divider `json:"divider"`
}

type DownloadableFileBlock ¶ added in v1.13.3

type DownloadableFileBlock interface {
	Block
	GetURL() string
	GetExpiryTime() *time.Time
}

DownloadableFileBlock is an interface for blocks that can be downloaded such as Pdf, FileBlock, and Image
type DualProperty ¶ added in v1.12.2

type DualProperty struct{}

type EmailProperty ¶

type EmailProperty struct {
	ID    PropertyID   `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Email string       `json:"email"`
}

func (EmailProperty) GetID ¶ added in v1.12.9

func (p EmailProperty) GetID() string

func (EmailProperty) GetType ¶

func (p EmailProperty) GetType() PropertyType

type EmailPropertyConfig ¶ added in v1.2.0

type EmailPropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Email struct{}           `json:"email"`
}

func (EmailPropertyConfig) GetID ¶ added in v1.13.0

func (p EmailPropertyConfig) GetID() PropertyID

func (EmailPropertyConfig) GetType ¶ added in v1.2.0

func (p EmailPropertyConfig) GetType() PropertyConfigType

type Embed ¶ added in v1.5.0

type Embed struct {
	Caption []RichText `json:"caption,omitempty"`
	URL     string     `json:"url"`
}

type EmbedBlock ¶ added in v1.5.0

type EmbedBlock struct {
	BasicBlock
	Embed Embed `json:"embed"`
}

func (EmbedBlock) GetRichTextString ¶ added in v1.12.10

func (b EmbedBlock) GetRichTextString() string

type Emoji ¶ added in v1.5.3

type Emoji string

type Equation ¶ added in v1.7.0

type Equation struct {
	Expression string `json:"expression"`
}

type EquationBlock ¶ added in v1.7.0

type EquationBlock struct {
	BasicBlock
	Equation Equation `json:"equation"`
}

func (EquationBlock) GetRichTextString ¶ added in v1.12.10

func (b EquationBlock) GetRichTextString() string

type Error ¶

type Error struct {
	Object  ObjectType `json:"object"`
	Status  int        `json:"status"`
	Code    ErrorCode  `json:"code"`
	Message string     `json:"message"`
}

func (*Error) Error ¶

func (e *Error) Error() string

type ErrorCode ¶

type ErrorCode string

type ExternalAccount ¶ added in v1.12.0

type ExternalAccount struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type File ¶ added in v1.2.0

type File struct {
	Name     string      `json:"name"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type FileBlock ¶ added in v1.5.0

type FileBlock struct {
	BasicBlock
	File BlockFile `json:"file"`
}

func (*FileBlock) GetExpiryTime ¶ added in v1.13.3

func (b *FileBlock) GetExpiryTime() *time.Time

GetExpiryTime implements DownloadableFileBlock interface for FileBlock
func (FileBlock) GetRichTextString ¶ added in v1.12.10

func (b FileBlock) GetRichTextString() string

func (*FileBlock) GetURL ¶ added in v1.13.3

func (b *FileBlock) GetURL() string

GetURL implements DownloadableFileBlock interface for FileBlock
type FileObject ¶ added in v1.5.0

type FileObject struct {
	URL        string     `json:"url,omitempty"`
	ExpiryTime *time.Time `json:"expiry_time,omitempty"`
}

type FileType ¶ added in v1.5.0

type FileType string

const (
	FileTypeFile     FileType = "file"
	FileTypeExternal FileType = "external"
)

type FilesFilterCondition ¶ added in v1.2.0

type FilesFilterCondition struct {
	IsEmpty    bool `json:"is_empty,omitempty"`
	IsNotEmpty bool `json:"is_not_empty,omitempty"`
}

type FilesProperty ¶ added in v1.2.0

type FilesProperty struct {
	ID    ObjectID     `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Files []File       `json:"files"`
}

func (FilesProperty) GetID ¶ added in v1.12.9

func (p FilesProperty) GetID() string

func (FilesProperty) GetType ¶ added in v1.2.0

func (p FilesProperty) GetType() PropertyType

type FilesPropertyConfig ¶ added in v1.2.0

type FilesPropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Files struct{}           `json:"files"`
}

func (FilesPropertyConfig) GetID ¶ added in v1.13.0

func (p FilesPropertyConfig) GetID() PropertyID

func (FilesPropertyConfig) GetType ¶ added in v1.2.0

func (p FilesPropertyConfig) GetType() PropertyConfigType

type Filter ¶

type Filter interface {
	// contains filtered or unexported methods
}

type FilterOperator ¶

type FilterOperator string

const (
	FilterOperatorAND FilterOperator = "and"
	FilterOperatorOR  FilterOperator = "or"
)

type FormatType ¶

type FormatType string

const (
	FormatNumber           FormatType = "number"
	FormatNumberWithCommas FormatType = "number_with_commas"
	FormatPercent          FormatType = "percent"
	FormatDollar           FormatType = "dollar"
	FormatCanadianDollar   FormatType = "canadian_dollar"
	FormatEuro             FormatType = "euro"
	FormatPound            FormatType = "pound"
	FormatYen              FormatType = "yen"
	FormatRuble            FormatType = "ruble"
	FormatRupee            FormatType = "rupee"
	FormatWon              FormatType = "won"
	FormatYuan             FormatType = "yuan"
	FormatReal             FormatType = "real"
	FormatLira             FormatType = "lira"
	FormatRupiah           FormatType = "rupiah"
	FormatFranc            FormatType = "franc"
	FormatHongKongDollar   FormatType = "hong_kong_dollar"
	FormatNewZealandDollar FormatType = "hong_kong_dollar"
	FormatKrona            FormatType = "krona"
	FormatNorwegianKrone   FormatType = "norwegian_krone"
	FormatMexicanPeso      FormatType = "mexican_peso"
	FormatRand             FormatType = "rand"
	FormatNewTaiwanDollar  FormatType = "new_taiwan_dollar"
	FormatDanishKrone      FormatType = "danish_krone"
	FormatZloty            FormatType = "zloty"
	FormatBath             FormatType = "baht"
	FormatForint           FormatType = "forint"
	FormatKoruna           FormatType = "koruna"
	FormatShekel           FormatType = "shekel"
	FormatChileanPeso      FormatType = "chilean_peso"
	FormatPhilippinePeso   FormatType = "philippine_peso"
	FormatDirham           FormatType = "dirham"
	FormatColombianPeso    FormatType = "colombian_peso"
	FormatRiyal            FormatType = "riyal"
	FormatRinggit          FormatType = "ringgit"
	FormatLeu              FormatType = "leu"
	FormatArgentinePeso    FormatType = "argentine_peso"
	FormatUruguayanPeso    FormatType = "uruguayan_peso"
	FormatSingaporeDollar  FormatType = "singapore_dollar"
)

func (FormatType) String ¶

func (ft FormatType) String() string

type Formula ¶ added in v1.2.0

type Formula struct {
	Type    FormulaType `json:"type,omitempty"`
	String  string      `json:"string,omitempty"`
	Number  float64     `json:"number,omitempty"`
	Boolean bool        `json:"boolean,omitempty"`
	Date    *DateObject `json:"date,omitempty"`
}

type FormulaConfig ¶ added in v1.2.0

type FormulaConfig struct {
	Expression string `json:"expression"`
}

type FormulaFilterCondition ¶ added in v1.2.0

type FormulaFilterCondition struct {
	// DEPRECATED use `String` instead
	Text     *TextFilterCondition     `json:"text,omitempty"`
	String   *TextFilterCondition     `json:"string,omitempty"`
	Checkbox *CheckboxFilterCondition `json:"checkbox,omitempty"`
	Number   *NumberFilterCondition   `json:"number,omitempty"`
	Date     *DateFilterCondition     `json:"date,omitempty"`
}

type FormulaProperty ¶

type FormulaProperty struct {
	ID      ObjectID     `json:"id,omitempty"`
	Type    PropertyType `json:"type,omitempty"`
	Formula Formula      `json:"formula"`
}

func (FormulaProperty) GetID ¶ added in v1.12.9

func (p FormulaProperty) GetID() string

func (FormulaProperty) GetType ¶

func (p FormulaProperty) GetType() PropertyType

type FormulaPropertyConfig ¶ added in v1.2.0

type FormulaPropertyConfig struct {
	ID      PropertyID         `json:"id,omitempty"`
	Type    PropertyConfigType `json:"type"`
	Formula FormulaConfig      `json:"formula"`
}

func (FormulaPropertyConfig) GetID ¶ added in v1.13.0

func (p FormulaPropertyConfig) GetID() PropertyID

func (FormulaPropertyConfig) GetType ¶ added in v1.2.0

func (p FormulaPropertyConfig) GetType() PropertyConfigType

type FormulaType ¶ added in v1.2.0

type FormulaType string

const (
	FormulaTypeString  FormulaType = "string"
	FormulaTypeNumber  FormulaType = "number"
	FormulaTypeBoolean FormulaType = "boolean"
	FormulaTypeDate    FormulaType = "date"
)

type FunctionType ¶

type FunctionType string

const (
	FunctionCountAll          FunctionType = "count_all"
	FunctionCountValues       FunctionType = "count_values"
	FunctionCountUniqueValues FunctionType = "count_unique_values"
	FunctionCountEmpty        FunctionType = "count_empty"
	FunctionCountNotEmpty     FunctionType = "count_not_empty"
	FunctionPercentEmpty      FunctionType = "percent_empty"
	FunctionPercentNotEmpty   FunctionType = "percent_not_empty"
	FunctionSum               FunctionType = "sum"
	FunctionAverage           FunctionType = "average"
	FunctionMedian            FunctionType = "median"
	FunctionMin               FunctionType = "min"
	FunctionMax               FunctionType = "max"
	FunctionRange             FunctionType = "range"
)

func (FunctionType) String ¶

func (ft FunctionType) String() string

type GetChildrenResponse ¶

type GetChildrenResponse struct {
	Object     ObjectType `json:"object"`
	Results    Blocks     `json:"results"`
	NextCursor string     `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}

type GroupConfig ¶ added in v1.12.9

type GroupConfig struct {
	ID        ObjectID   `json:"id"`
	Name      string     `json:"name"`
	Color     string     `json:"color"`
	OptionIDs []ObjectID `json:"option_ids"`
}

type Heading ¶ added in v1.4.0

type Heading struct {
	RichText     []RichText `json:"rich_text"`
	Children     Blocks     `json:"children,omitempty"`
	Color        string     `json:"color,omitempty"`
	IsToggleable bool       `json:"is_toggleable,omitempty"`
}

type Heading1Block ¶

type Heading1Block struct {
	BasicBlock
	Heading1 Heading `json:"heading_1"`
}

func (Heading1Block) GetRichTextString ¶ added in v1.12.10

func (h Heading1Block) GetRichTextString() string

type Heading2Block ¶

type Heading2Block struct {
	BasicBlock
	Heading2 Heading `json:"heading_2"`
}

func (Heading2Block) GetRichTextString ¶ added in v1.12.10

func (h Heading2Block) GetRichTextString() string

type Heading3Block ¶

type Heading3Block struct {
	BasicBlock
	Heading3 Heading `json:"heading_3"`
}

func (Heading3Block) GetRichTextString ¶ added in v1.12.10

func (h Heading3Block) GetRichTextString() string

type Icon ¶ added in v1.5.3

type Icon struct {
	Type     FileType    `json:"type"`
	Emoji    *Emoji      `json:"emoji,omitempty"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

func (Icon) GetURL ¶ added in v1.5.3

func (i Icon) GetURL() string

GetURL returns the external or internal URL depending on the image type.
type Image ¶ added in v1.5.0

type Image struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type,omitempty"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

func (Image) GetURL ¶ added in v1.5.3

func (i Image) GetURL() string

GetURL returns the external or internal URL depending on the image type.
type ImageBlock ¶ added in v1.5.0

type ImageBlock struct {
	BasicBlock
	Image Image `json:"image"`
}

func (*ImageBlock) GetExpiryTime ¶ added in v1.13.3

func (b *ImageBlock) GetExpiryTime() *time.Time

GetExpiryTime implements DownloadableFileBlock interface for ImageBlock
func (ImageBlock) GetRichTextString ¶ added in v1.12.10

func (b ImageBlock) GetRichTextString() string

func (*ImageBlock) GetURL ¶ added in v1.13.3

func (b *ImageBlock) GetURL() string

GetURL implements DownloadableFileBlock interface for ImageBlock
type LastEditedByProperty ¶

type LastEditedByProperty struct {
	ID           ObjectID     `json:"id,omitempty"`
	Type         PropertyType `json:"type,omitempty"`
	LastEditedBy User         `json:"last_edited_by"`
}

func (LastEditedByProperty) GetID ¶ added in v1.12.9

func (p LastEditedByProperty) GetID() string

func (LastEditedByProperty) GetType ¶

func (p LastEditedByProperty) GetType() PropertyType

type LastEditedByPropertyConfig ¶ added in v1.2.0

type LastEditedByPropertyConfig struct {
	ID           PropertyID         `json:"id"`
	Type         PropertyConfigType `json:"type"`
	LastEditedBy struct{}           `json:"last_edited_by"`
}

func (LastEditedByPropertyConfig) GetID ¶ added in v1.13.0

func (p LastEditedByPropertyConfig) GetID() PropertyID

func (LastEditedByPropertyConfig) GetType ¶ added in v1.2.0

func (p LastEditedByPropertyConfig) GetType() PropertyConfigType

type LastEditedTimeProperty ¶

type LastEditedTimeProperty struct {
	ID             ObjectID     `json:"id,omitempty"`
	Type           PropertyType `json:"type,omitempty"`
	LastEditedTime time.Time    `json:"last_edited_time"`
}

func (LastEditedTimeProperty) GetID ¶ added in v1.12.9

func (p LastEditedTimeProperty) GetID() string

func (LastEditedTimeProperty) GetType ¶

func (p LastEditedTimeProperty) GetType() PropertyType

type LastEditedTimePropertyConfig ¶ added in v1.2.0

type LastEditedTimePropertyConfig struct {
	ID             PropertyID         `json:"id"`
	Type           PropertyConfigType `json:"type"`
	LastEditedTime struct{}           `json:"last_edited_time"`
}

func (LastEditedTimePropertyConfig) GetID ¶ added in v1.13.0

func (p LastEditedTimePropertyConfig) GetID() PropertyID

func (LastEditedTimePropertyConfig) GetType ¶ added in v1.2.0

func (p LastEditedTimePropertyConfig) GetType() PropertyConfigType

type Link ¶ added in v1.5.1

type Link struct {
	Url string `json:"url,omitempty"`
}

type LinkPreview ¶ added in v1.7.0

type LinkPreview struct {
	URL string `json:"url"`
}

type LinkPreviewBlock ¶ added in v1.7.0

type LinkPreviewBlock struct {
	BasicBlock
	LinkPreview LinkPreview `json:"link_preview"`
}

NOTE: will only be returned by the API. Cannot be created by the API. https://developers.notion.com/reference/block#link-preview-blocks
func (LinkPreviewBlock) GetRichTextString ¶ added in v1.12.10

func (b LinkPreviewBlock) GetRichTextString() string

type LinkToPage ¶ added in v1.7.0

type LinkToPage struct {
	Type       BlockType  `json:"type"`
	PageID     PageID     `json:"page_id,omitempty"`
	DatabaseID DatabaseID `json:"database_id,omitempty"`
}

type LinkToPageBlock ¶ added in v1.7.0

type LinkToPageBlock struct {
	BasicBlock
	LinkToPage LinkToPage `json:"link_to_page"`
}

type ListItem ¶ added in v1.4.0

type ListItem struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type Mention ¶ added in v1.9.1

type Mention struct {
	Type            MentionType      `json:"type,omitempty"`
	Database        *DatabaseMention `json:"database,omitempty"`
	Page            *PageMention     `json:"page,omitempty"`
	User            *User            `json:"user,omitempty"`
	Date            *DateObject      `json:"date,omitempty"`
	TemplateMention *TemplateMention `json:"template_mention,omitempty"`
}

type MentionType ¶ added in v1.9.1

type MentionType string

const (
	MentionTypeDatabase        MentionType = "database"
	MentionTypePage            MentionType = "page"
	MentionTypeUser            MentionType = "user"
	MentionTypeDate            MentionType = "date"
	MentionTypeTemplateMention MentionType = "template_mention"
)

func (MentionType) String ¶ added in v1.9.1

func (mType MentionType) String() string

type MultiSelectFilterCondition ¶ added in v1.2.0

type MultiSelectFilterCondition struct {
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type MultiSelectProperty ¶

type MultiSelectProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	MultiSelect []Option     `json:"multi_select"`
}

func (MultiSelectProperty) GetID ¶ added in v1.12.9

func (p MultiSelectProperty) GetID() string

func (MultiSelectProperty) GetType ¶

func (p MultiSelectProperty) GetType() PropertyType

type MultiSelectPropertyConfig ¶ added in v1.2.0

type MultiSelectPropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	MultiSelect Select             `json:"multi_select"`
}

func (MultiSelectPropertyConfig) GetID ¶ added in v1.13.0

func (p MultiSelectPropertyConfig) GetID() PropertyID

func (MultiSelectPropertyConfig) GetType ¶ added in v1.2.0

func (p MultiSelectPropertyConfig) GetType() PropertyConfigType

type NumberFilterCondition ¶ added in v1.2.0

type NumberFilterCondition struct {
	Equals               *float64 `json:"equals,omitempty"`
	DoesNotEqual         *float64 `json:"does_not_equal,omitempty"`
	GreaterThan          *float64 `json:"greater_than,omitempty"`
	LessThan             *float64 `json:"less_than,omitempty"`
	GreaterThanOrEqualTo *float64 `json:"greater_than_or_equal_to,omitempty"`
	LessThanOrEqualTo    *float64 `json:"less_than_or_equal_to,omitempty"`
	IsEmpty              bool     `json:"is_empty,omitempty"`
	IsNotEmpty           bool     `json:"is_not_empty,omitempty"`
}

type NumberFormat ¶ added in v1.8.6

type NumberFormat struct {
	Format FormatType `json:"format"`
}

type NumberProperty ¶

type NumberProperty struct {
	ID     PropertyID   `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Number float64      `json:"number"`
}

func (NumberProperty) GetID ¶ added in v1.12.9

func (p NumberProperty) GetID() string

func (NumberProperty) GetType ¶

func (p NumberProperty) GetType() PropertyType

type NumberPropertyConfig ¶ added in v1.2.0

type NumberPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Number NumberFormat       `json:"number"`
}

func (NumberPropertyConfig) GetID ¶ added in v1.13.0

func (p NumberPropertyConfig) GetID() PropertyID

func (NumberPropertyConfig) GetType ¶ added in v1.2.0

func (p NumberPropertyConfig) GetType() PropertyConfigType

type NumberedListItemBlock ¶

type NumberedListItemBlock struct {
	BasicBlock
	NumberedListItem ListItem `json:"numbered_list_item"`
}

func (NumberedListItemBlock) GetRichTextString ¶ added in v1.12.10

func (n NumberedListItemBlock) GetRichTextString() string

type Object ¶

type Object interface {
	GetObject() ObjectType
}

type ObjectID ¶

type ObjectID string

func (ObjectID) String ¶

func (oID ObjectID) String() string

type ObjectType ¶

type ObjectType string

const (
	ObjectTypeDatabase ObjectType = "database"
	ObjectTypeBlock    ObjectType = "block"
	ObjectTypePage     ObjectType = "page"
	ObjectTypeList     ObjectType = "list"
	ObjectTypeText     ObjectType = "text"
	ObjectTypeUser     ObjectType = "user"
	ObjectTypeError    ObjectType = "error"
	ObjectTypeComment  ObjectType = "comment"
)

func (ObjectType) String ¶

func (ot ObjectType) String() string

type Option ¶

type Option struct {
	ID    PropertyID `json:"id,omitempty"`
	Name  string     `json:"name"`
	Color Color      `json:"color,omitempty"`
}

type OrCompoundFilter ¶ added in v1.8.5

type OrCompoundFilter []Filter

func (OrCompoundFilter) MarshalJSON ¶ added in v1.8.5

func (f OrCompoundFilter) MarshalJSON() ([]byte, error)

type Owner ¶ added in v1.10.2

type Owner struct {
	Type      string `json:"type"`
	Workspace bool   `json:"workspace"`
}

type Page ¶

type Page struct {
	Object         ObjectType `json:"object"`
	ID             ObjectID   `json:"id"`
	CreatedTime    time.Time  `json:"created_time"`
	LastEditedTime time.Time  `json:"last_edited_time"`
	CreatedBy      User       `json:"created_by,omitempty"`
	LastEditedBy   User       `json:"last_edited_by,omitempty"`
	Archived       bool       `json:"archived"`
	Properties     Properties `json:"properties"`
	Parent         Parent     `json:"parent"`
	URL            string     `json:"url"`
	PublicURL      string     `json:"public_url"`
	Icon           *Icon      `json:"icon,omitempty"`
	Cover          *Image     `json:"cover,omitempty"`
}

The Page object contains the page property values of a single Notion page.

See https://developers.notion.com/reference/page
func (*Page) GetObject ¶

func (p *Page) GetObject() ObjectType

type PageClient ¶

type PageClient struct {
	// contains filtered or unexported fields
}

func (*PageClient) Create ¶

func (pc *PageClient) Create(ctx context.Context, requestBody *PageCreateRequest) (*Page, error)

Creates a new page that is a child of an existing page or database.

If the new page is a child of an existing page,title is the only valid property in the properties body param.

If the new page is a child of an existing database, the keys of the properties object body param must match the parent database's properties.

This endpoint can be used to create a new page with or without content using the children option. To add content to a page after creating it, use the Append block children endpoint.

Returns a new page object.

See https://developers.notion.com/reference/post-page
func (*PageClient) Get ¶

func (pc *PageClient) Get(ctx context.Context, id PageID) (*Page, error)

Retrieves a Page object using the ID specified.

Responses contains page properties, not page content. To fetch page content, use the Retrieve block children endpoint.

Page properties are limited to up to 25 references per page property. To retrieve data related to properties that have more than 25 references, use the Retrieve a page property endpoint.

See https://developers.notion.com/reference/get-page
func (*PageClient) Update ¶

func (pc *PageClient) Update(ctx context.Context, id PageID, request *PageUpdateRequest) (*Page, error)

Updates the properties of a page in a database. The properties body param of this endpoint can only be used to update the properties of a page that is a child of a database. The page’s properties schema must match the parent database’s properties.

This endpoint can be used to update any page icon or cover, and can be used to archive or restore any page.

To add page content instead of page properties, use the append block children endpoint. The page_id can be passed as the block_id when adding block children to the page.

Returns the updated page object.

See https://developers.notion.com/reference/patch-page
type PageCreateRequest ¶

type PageCreateRequest struct {
	// The parent page or database where the new page is inserted, represented as
	// a JSON object with a page_id or database_id key, and the corresponding ID.
	Parent Parent `json:"parent"`
	// The values of the page’s properties. If the parent is a database, then the
	// schema must match the parent database’s properties. If the parent is a page,
	// then the only valid object key is title.
	Properties Properties `json:"properties"`
	// The content to be rendered on the new page, represented as an array of
	// block objects.
	Children []Block `json:"children,omitempty"`
	// The icon of the new page. Either an emoji object or an external file object.
	Icon *Icon `json:"icon,omitempty"`
	// The cover image of the new page, represented as a file object.
	Cover *Image `json:"cover,omitempty"`
}

PageCreateRequest represents the request body for PageClient.Create.
type PageID ¶

type PageID string

func (PageID) String ¶

func (pID PageID) String() string

type PageMention ¶ added in v1.9.1

type PageMention struct {
	ID ObjectID `json:"id"`
}

type PageService ¶

type PageService interface {
	Create(context.Context, *PageCreateRequest) (*Page, error)
	Get(context.Context, PageID) (*Page, error)
	Update(context.Context, PageID, *PageUpdateRequest) (*Page, error)
}

type PageUpdateRequest ¶

type PageUpdateRequest struct {
	// The property values to update for the page. The keys are the names or IDs
	// of the property and the values are property values. If a page property ID
	// is not included, then it is not changed.
	Properties Properties `json:"properties,omitempty"`
	// Whether the page is archived (deleted). Set to true to archive a page. Set
	// to false to un-archive (restore) a page.
	Archived bool `json:"archived"`
	// A page icon for the page. Supported types are external file object or emoji
	// object.
	Icon *Icon `json:"icon,omitempty"`
	// A cover image for the page. Only external file objects are supported.
	Cover *Image `json:"cover,omitempty"`
}

PageUpdateRequest represents the request body for PageClient.Update.
type Pagination ¶

type Pagination struct {
	StartCursor Cursor
	PageSize    int
}

func (*Pagination) ToQuery ¶

func (p *Pagination) ToQuery() map[string]string

type Paragraph ¶

type Paragraph struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type ParagraphBlock ¶

type ParagraphBlock struct {
	BasicBlock
	Paragraph Paragraph `json:"paragraph"`
}

func (ParagraphBlock) GetRichTextString ¶ added in v1.12.10

func (p ParagraphBlock) GetRichTextString() string

type Parent ¶

type Parent struct {
	Type       ParentType `json:"type,omitempty"`
	PageID     PageID     `json:"page_id,omitempty"`
	DatabaseID DatabaseID `json:"database_id,omitempty"`
	BlockID    BlockID    `json:"block_id,omitempty"`
	Workspace  bool       `json:"workspace,omitempty"`
}

Pages, databases, and blocks are either located inside other pages, databases, and blocks, or are located at the top level of a workspace. This location is known as the "parent". Parent information is represented by a consistent parent object throughout the API.

See https://developers.notion.com/reference/parent-object
type ParentType ¶

type ParentType string

const (
	ParentTypeDatabaseID ParentType = "database_id"
	ParentTypePageID     ParentType = "page_id"
	ParentTypeWorkspace  ParentType = "workspace"
	ParentTypeBlockID    ParentType = "block_id"
)

type Pdf ¶ added in v1.5.0

type Pdf struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type,omitempty"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type PdfBlock ¶ added in v1.5.0

type PdfBlock struct {
	BasicBlock
	Pdf Pdf `json:"pdf"`
}

func (*PdfBlock) GetExpiryTime ¶ added in v1.13.3

func (b *PdfBlock) GetExpiryTime() *time.Time

GetExpiryTime implements DownloadableFileBlock interface for PdfBlock
func (PdfBlock) GetRichTextString ¶ added in v1.12.10

func (b PdfBlock) GetRichTextString() string

func (*PdfBlock) GetURL ¶ added in v1.13.3

func (b *PdfBlock) GetURL() string

GetURL implements DownloadableFileBlock interface for PdfBlock
type PeopleFilterCondition ¶ added in v1.2.0

type PeopleFilterCondition struct {
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type PeopleProperty ¶

type PeopleProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	People []User       `json:"people"`
}

func (PeopleProperty) GetID ¶ added in v1.12.9

func (p PeopleProperty) GetID() string

func (PeopleProperty) GetType ¶

func (p PeopleProperty) GetType() PropertyType

type PeoplePropertyConfig ¶ added in v1.2.0

type PeoplePropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	People struct{}           `json:"people"`
}

func (PeoplePropertyConfig) GetID ¶ added in v1.13.0

func (p PeoplePropertyConfig) GetID() PropertyID

func (PeoplePropertyConfig) GetType ¶ added in v1.2.0

func (p PeoplePropertyConfig) GetType() PropertyConfigType

type Person ¶

type Person struct {
	Email string `json:"email"`
}

type PhoneNumberProperty ¶

type PhoneNumberProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	PhoneNumber string       `json:"phone_number"`
}

func (PhoneNumberProperty) GetID ¶ added in v1.12.9

func (p PhoneNumberProperty) GetID() string

func (PhoneNumberProperty) GetType ¶

func (p PhoneNumberProperty) GetType() PropertyType

type PhoneNumberPropertyConfig ¶ added in v1.2.0

type PhoneNumberPropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	PhoneNumber struct{}           `json:"phone_number"`
}

func (PhoneNumberPropertyConfig) GetID ¶ added in v1.13.0

func (p PhoneNumberPropertyConfig) GetID() PropertyID

func (PhoneNumberPropertyConfig) GetType ¶ added in v1.2.0

func (p PhoneNumberPropertyConfig) GetType() PropertyConfigType

type Properties ¶

type Properties map[string]Property

func (*Properties) UnmarshalJSON ¶

func (p *Properties) UnmarshalJSON(data []byte) error

type Property ¶

type Property interface {
	GetID() string
	GetType() PropertyType
}

type PropertyArray ¶ added in v1.7.1

type PropertyArray []Property

func (*PropertyArray) UnmarshalJSON ¶ added in v1.7.1

func (arr *PropertyArray) UnmarshalJSON(data []byte) error

type PropertyConfig ¶ added in v1.2.0

type PropertyConfig interface {
	GetType() PropertyConfigType
	GetID() PropertyID
}

type PropertyConfigType ¶ added in v1.2.0

type PropertyConfigType string

const (
	PropertyConfigTypeTitle       PropertyConfigType = "title"
	PropertyConfigTypeRichText    PropertyConfigType = "rich_text"
	PropertyConfigTypeNumber      PropertyConfigType = "number"
	PropertyConfigTypeSelect      PropertyConfigType = "select"
	PropertyConfigTypeMultiSelect PropertyConfigType = "multi_select"
	PropertyConfigTypeDate        PropertyConfigType = "date"
	PropertyConfigTypePeople      PropertyConfigType = "people"
	PropertyConfigTypeFiles       PropertyConfigType = "files"
	PropertyConfigTypeCheckbox    PropertyConfigType = "checkbox"
	PropertyConfigTypeURL         PropertyConfigType = "url"
	PropertyConfigTypeEmail       PropertyConfigType = "email"
	PropertyConfigTypePhoneNumber PropertyConfigType = "phone_number"
	PropertyConfigTypeFormula     PropertyConfigType = "formula"
	PropertyConfigTypeRelation    PropertyConfigType = "relation"
	PropertyConfigTypeRollup      PropertyConfigType = "rollup"
	PropertyConfigCreatedTime     PropertyConfigType = "created_time"
	PropertyConfigCreatedBy       PropertyConfigType = "created_by"
	PropertyConfigLastEditedTime  PropertyConfigType = "last_edited_time"
	PropertyConfigLastEditedBy    PropertyConfigType = "last_edited_by"
	PropertyConfigStatus          PropertyConfigType = "status"
	PropertyConfigUniqueID        PropertyConfigType = "unique_id"
	PropertyConfigVerification    PropertyConfigType = "verification"
	PropertyConfigButton          PropertyConfigType = "button"
)

type PropertyConfigs ¶ added in v1.2.0

type PropertyConfigs map[string]PropertyConfig

func (*PropertyConfigs) UnmarshalJSON ¶ added in v1.2.0

func (p *PropertyConfigs) UnmarshalJSON(data []byte) error

type PropertyFilter ¶

type PropertyFilter struct {
	Property    string                      `json:"property"`
	RichText    *TextFilterCondition        `json:"rich_text,omitempty"`
	Number      *NumberFilterCondition      `json:"number,omitempty"`
	Checkbox    *CheckboxFilterCondition    `json:"checkbox,omitempty"`
	Select      *SelectFilterCondition      `json:"select,omitempty"`
	MultiSelect *MultiSelectFilterCondition `json:"multi_select,omitempty"`
	Date        *DateFilterCondition        `json:"date,omitempty"`
	People      *PeopleFilterCondition      `json:"people,omitempty"`
	Files       *FilesFilterCondition       `json:"files,omitempty"`
	Relation    *RelationFilterCondition    `json:"relation,omitempty"`
	Formula     *FormulaFilterCondition     `json:"formula,omitempty"`
	Rollup      *RollupFilterCondition      `json:"rollup,omitempty"`
	Status      *StatusFilterCondition      `json:"status,omitempty"`
	UniqueId    *UniqueIdFilterCondition    `json:"unique_id,omitempty"`
}

type PropertyID ¶

type PropertyID string

func (PropertyID) String ¶

func (pID PropertyID) String() string

type PropertyType ¶

type PropertyType string

const (
	PropertyTypeTitle          PropertyType = "title"
	PropertyTypeRichText       PropertyType = "rich_text"
	PropertyTypeText           PropertyType = "text"
	PropertyTypeNumber         PropertyType = "number"
	PropertyTypeSelect         PropertyType = "select"
	PropertyTypeMultiSelect    PropertyType = "multi_select"
	PropertyTypeDate           PropertyType = "date"
	PropertyTypeFormula        PropertyType = "formula"
	PropertyTypeRelation       PropertyType = "relation"
	PropertyTypeRollup         PropertyType = "rollup"
	PropertyTypePeople         PropertyType = "people"
	PropertyTypeFiles          PropertyType = "files"
	PropertyTypeCheckbox       PropertyType = "checkbox"
	PropertyTypeURL            PropertyType = "url"
	PropertyTypeEmail          PropertyType = "email"
	PropertyTypePhoneNumber    PropertyType = "phone_number"
	PropertyTypeCreatedTime    PropertyType = "created_time"
	PropertyTypeCreatedBy      PropertyType = "created_by"
	PropertyTypeLastEditedTime PropertyType = "last_edited_time"
	PropertyTypeLastEditedBy   PropertyType = "last_edited_by"
	PropertyTypeStatus         PropertyType = "status"
	PropertyTypeUniqueID       PropertyType = "unique_id"
	PropertyTypeVerification   PropertyType = "verification"
	PropertyTypeButton         PropertyType = "button"
)

type Quote ¶ added in v1.5.3

type Quote struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type QuoteBlock ¶ added in v1.5.3

type QuoteBlock struct {
	BasicBlock
	Quote Quote `json:"quote"`
}

func (QuoteBlock) GetRichTextString ¶ added in v1.12.10

func (q QuoteBlock) GetRichTextString() string

type RateLimitedError ¶ added in v1.10.1

type RateLimitedError struct {
	Message string
}

func (*RateLimitedError) Error ¶ added in v1.10.1

func (e *RateLimitedError) Error() string

type Relation ¶

type Relation struct {
	ID PageID `json:"id"`
}

type RelationConfig ¶ added in v1.2.0

type RelationConfig struct {
	DatabaseID         DatabaseID         `json:"database_id"`
	SyncedPropertyID   PropertyID         `json:"synced_property_id,omitempty"`
	SyncedPropertyName string             `json:"synced_property_name,omitempty"`
	Type               RelationConfigType `json:"type,omitempty"`
	SingleProperty     *SingleProperty    `json:"single_property,omitempty"`
	DualProperty       *DualProperty      `json:"dual_property,omitempty"`
}

type RelationConfigType ¶ added in v1.12.2

type RelationConfigType string

const (
	RelationSingleProperty RelationConfigType = "single_property"
	RelationDualProperty   RelationConfigType = "dual_property"
)

func (RelationConfigType) String ¶ added in v1.12.2

func (rp RelationConfigType) String() string

type RelationFilterCondition ¶ added in v1.2.0

type RelationFilterCondition struct {
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type RelationObject ¶

type RelationObject struct {
	Database           DatabaseID `json:"database"`
	SyncedPropertyName string     `json:"synced_property_name"`
}

type RelationProperty ¶

type RelationProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	Relation []Relation   `json:"relation"`
}

func (RelationProperty) GetID ¶ added in v1.12.9

func (p RelationProperty) GetID() string

func (RelationProperty) GetType ¶

func (p RelationProperty) GetType() PropertyType

type RelationPropertyConfig ¶ added in v1.2.0

type RelationPropertyConfig struct {
	Type     PropertyConfigType `json:"type"`
	Relation RelationConfig     `json:"relation"`
}

func (RelationPropertyConfig) GetID ¶ added in v1.13.0

func (p RelationPropertyConfig) GetID() PropertyID

func (RelationPropertyConfig) GetType ¶ added in v1.2.0

func (p RelationPropertyConfig) GetType() PropertyConfigType

type RichText ¶

type RichText struct {
	Type        ObjectType   `json:"type,omitempty"`
	Text        *Text        `json:"text,omitempty"`
	Mention     *Mention     `json:"mention,omitempty"`
	Equation    *Equation    `json:"equation,omitempty"`
	Annotations *Annotations `json:"annotations,omitempty"`
	PlainText   string       `json:"plain_text,omitempty"`
	Href        string       `json:"href,omitempty"`
}

type RichTextProperty ¶

type RichTextProperty struct {
	ID       PropertyID   `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	RichText []RichText   `json:"rich_text"`
}

func (RichTextProperty) GetID ¶ added in v1.12.9

func (p RichTextProperty) GetID() string

func (RichTextProperty) GetType ¶

func (p RichTextProperty) GetType() PropertyType

type RichTextPropertyConfig ¶ added in v1.2.0

type RichTextPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	RichText struct{}           `json:"rich_text"`
}

func (RichTextPropertyConfig) GetID ¶ added in v1.13.0

func (p RichTextPropertyConfig) GetID() PropertyID

func (RichTextPropertyConfig) GetType ¶ added in v1.2.0

func (p RichTextPropertyConfig) GetType() PropertyConfigType

type Rollup ¶

type Rollup struct {
	Type   RollupType    `json:"type,omitempty"`
	Number float64       `json:"number,omitempty"`
	Date   *DateObject   `json:"date,omitempty"`
	Array  PropertyArray `json:"array,omitempty"`
}

type RollupConfig ¶ added in v1.2.0

type RollupConfig struct {
	RelationPropertyName string       `json:"relation_property_name"`
	RelationPropertyID   PropertyID   `json:"relation_property_id"`
	RollupPropertyName   string       `json:"rollup_property_name"`
	RollupPropertyID     PropertyID   `json:"rollup_property_id"`
	Function             FunctionType `json:"function"`
}

type RollupFilterCondition ¶ added in v1.9.0

type RollupFilterCondition struct {
	Any    *RollupSubfilterCondition `json:"any,omitempty"`
	None   *RollupSubfilterCondition `json:"none,omitempty"`
	Every  *RollupSubfilterCondition `json:"every,omitempty"`
	Date   *DateFilterCondition      `json:"date,omitempty"`
	Number *NumberFilterCondition    `json:"number,omitempty"`
}

type RollupProperty ¶

type RollupProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Rollup Rollup       `json:"rollup"`
}

func (RollupProperty) GetID ¶ added in v1.12.9

func (p RollupProperty) GetID() string

func (RollupProperty) GetType ¶

func (p RollupProperty) GetType() PropertyType

type RollupPropertyConfig ¶ added in v1.2.0

type RollupPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Rollup RollupConfig       `json:"rollup"`
}

func (RollupPropertyConfig) GetID ¶ added in v1.13.0

func (p RollupPropertyConfig) GetID() PropertyID

func (RollupPropertyConfig) GetType ¶ added in v1.2.0

func (p RollupPropertyConfig) GetType() PropertyConfigType

type RollupSubfilterCondition ¶ added in v1.9.0

type RollupSubfilterCondition struct {
	RichText    *TextFilterCondition        `json:"rich_text,omitempty"`
	Number      *NumberFilterCondition      `json:"number,omitempty"`
	Checkbox    *CheckboxFilterCondition    `json:"checkbox,omitempty"`
	Select      *SelectFilterCondition      `json:"select,omitempty"`
	MultiSelect *MultiSelectFilterCondition `json:"multiSelect,omitempty"`
	Relation    *RelationFilterCondition    `json:"relation,omitempty"`
	Date        *DateFilterCondition        `json:"date,omitempty"`
	People      *PeopleFilterCondition      `json:"people,omitempty"`
	Files       *FilesFilterCondition       `json:"files,omitempty"`
}

type RollupType ¶ added in v1.2.0

type RollupType string

const (
	RollupTypeNumber RollupType = "number"
	RollupTypeDate   RollupType = "date"
	RollupTypeArray  RollupType = "array"
)

type SearchClient ¶

type SearchClient struct {
	// contains filtered or unexported fields
}

func (*SearchClient) Do ¶

func (sc *SearchClient) Do(ctx context.Context, request *SearchRequest) (*SearchResponse, error)

To limit the request to search only pages or to search only databases, use the filter param.

See https://developers.notion.com/reference/post-search
type SearchFilter ¶ added in v1.11.0

type SearchFilter struct {
	Value    string `json:"value"`
	Property string `json:"property"`
}

type SearchRequest ¶

type SearchRequest struct {
	// The text that the API compares page and database titles against.
	Query string `json:"query,omitempty"`
	// A set of criteria, direction and timestamp keys, that orders the results.
	// The only supported timestamp value is "last_edited_time". Supported
	// direction values are "ascending" and "descending". If sort is not provided,
	// then the most recently edited results are returned first.
	Sort *SortObject `json:"sort,omitempty"`
	// A set of criteria, value and property keys, that limits the results to
	// either only pages or only databases. Possible value values are "page" or
	// "database". The only supported property value is "object".
	Filter SearchFilter `json:"filter,omitempty"`
	// A cursor value returned in a previous response that If supplied, limits the
	// response to results starting after the cursor. If not supplied, then the
	// first page of results is returned. Refer to pagination for more details.
	StartCursor Cursor `json:"start_cursor,omitempty"`
	// The number of items from the full list to include in the response. Maximum: 100.
	PageSize int `json:"page_size,omitempty"`
}

type SearchResponse ¶

type SearchResponse struct {
	Object     ObjectType `json:"object"`
	Results    []Object   `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

func (*SearchResponse) UnmarshalJSON ¶

func (sr *SearchResponse) UnmarshalJSON(data []byte) error

type SearchService ¶

type SearchService interface {
	Do(context.Context, *SearchRequest) (*SearchResponse, error)
}

type Select ¶

type Select struct {
	Options []Option `json:"options"`
}

type SelectFilterCondition ¶ added in v1.2.0

type SelectFilterCondition struct {
	Equals       string `json:"equals,omitempty"`
	DoesNotEqual string `json:"does_not_equal,omitempty"`
	IsEmpty      bool   `json:"is_empty,omitempty"`
	IsNotEmpty   bool   `json:"is_not_empty,omitempty"`
}

type SelectProperty ¶

type SelectProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Select Option       `json:"select"`
}

func (SelectProperty) GetID ¶ added in v1.12.9

func (p SelectProperty) GetID() string

func (SelectProperty) GetType ¶

func (p SelectProperty) GetType() PropertyType

type SelectPropertyConfig ¶ added in v1.2.0

type SelectPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Select Select             `json:"select"`
}

func (SelectPropertyConfig) GetID ¶ added in v1.13.0

func (p SelectPropertyConfig) GetID() PropertyID

func (SelectPropertyConfig) GetType ¶ added in v1.2.0

func (p SelectPropertyConfig) GetType() PropertyConfigType

type SingleProperty ¶ added in v1.12.2

type SingleProperty struct{}

type SortObject ¶

type SortObject struct {
	Property  string        `json:"property,omitempty"`
	Timestamp TimestampType `json:"timestamp,omitempty"`
	Direction SortOrder     `json:"direction,omitempty"`
}

type SortOrder ¶

type SortOrder string

const (
	SortOrderASC  SortOrder = "ascending"
	SortOrderDESC SortOrder = "descending"
)

type Status ¶ added in v1.9.0

type Status = Option

type StatusConfig ¶ added in v1.12.9

type StatusConfig struct {
	Options []Option      `json:"options"`
	Groups  []GroupConfig `json:"groups"`
}

type StatusFilterCondition ¶ added in v1.9.2

type StatusFilterCondition struct {
	Equals       string `json:"equals,omitempty"`
	DoesNotEqual string `json:"does_not_equal,omitempty"`
	IsEmpty      bool   `json:"is_empty,omitempty"`
	IsNotEmpty   bool   `json:"is_not_empty,omitempty"`
}

type StatusProperty ¶ added in v1.9.0

type StatusProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Status Status       `json:"status"`
}

func (StatusProperty) GetID ¶ added in v1.12.9

func (p StatusProperty) GetID() string

func (StatusProperty) GetType ¶ added in v1.9.0

func (p StatusProperty) GetType() PropertyType

type StatusPropertyConfig ¶ added in v1.9.0

type StatusPropertyConfig struct {
	ID     PropertyID         `json:"id"`
	Type   PropertyConfigType `json:"type"`
	Status StatusConfig       `json:"status"`
}

func (StatusPropertyConfig) GetID ¶ added in v1.13.0

func (p StatusPropertyConfig) GetID() PropertyID

func (StatusPropertyConfig) GetType ¶ added in v1.9.0

func (p StatusPropertyConfig) GetType() PropertyConfigType

type Synced ¶ added in v1.7.0

type Synced struct {
	// SyncedFrom is nil for the original block.
	SyncedFrom *SyncedFrom `json:"synced_from"`
	Children   Blocks      `json:"children,omitempty"`
}

type SyncedBlock ¶ added in v1.7.0

type SyncedBlock struct {
	BasicBlock
	SyncedBlock Synced `json:"synced_block"`
}

type SyncedFrom ¶ added in v1.7.0

type SyncedFrom struct {
	BlockID BlockID `json:"block_id"`
}

type Table ¶ added in v1.7.3

type Table struct {
	TableWidth      int    `json:"table_width"`
	HasColumnHeader bool   `json:"has_column_header"`
	HasRowHeader    bool   `json:"has_row_header"`
	Children        Blocks `json:"children,omitempty"`
}

type TableBlock ¶ added in v1.7.3

type TableBlock struct {
	BasicBlock
	Table Table `json:"table"`
}

type TableOfContents ¶ added in v1.7.0

type TableOfContents struct {
	// empty
	Color string `json:"color,omitempty"`
}

type TableOfContentsBlock ¶ added in v1.5.3

type TableOfContentsBlock struct {
	BasicBlock
	TableOfContents TableOfContents `json:"table_of_contents"`
}

type TableRow ¶ added in v1.7.3

type TableRow struct {
	Cells [][]RichText `json:"cells"`
}

type TableRowBlock ¶ added in v1.7.3

type TableRowBlock struct {
	BasicBlock
	TableRow TableRow `json:"table_row"`
}

type Template ¶ added in v1.7.0

type Template struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
}

type TemplateBlock ¶ added in v1.7.0

type TemplateBlock struct {
	BasicBlock
	Template Template `json:"template"`
}

func (TemplateBlock) GetRichTextString ¶ added in v1.12.10

func (b TemplateBlock) GetRichTextString() string

type TemplateMention ¶ added in v1.9.1

type TemplateMention struct {
	Type                TemplateMentionType `json:"type"`
	TemplateMentionUser string              `json:"template_mention_user,omitempty"`
	TemplateMentionDate string              `json:"template_mention_date,omitempty"`
}

type TemplateMentionType ¶ added in v1.9.1

type TemplateMentionType string

const (
	TemplateMentionTypeUser TemplateMentionType = "template_mention_user"
	TemplateMentionTypeDate TemplateMentionType = "template_mention_date"
)

func (TemplateMentionType) String ¶ added in v1.9.1

func (tMType TemplateMentionType) String() string

type Text ¶

type Text struct {
	Content string `json:"content"`
	Link    *Link  `json:"link,omitempty"`
}

type TextFilterCondition ¶ added in v1.2.0

type TextFilterCondition struct {
	Equals         string `json:"equals,omitempty"`
	DoesNotEqual   string `json:"does_not_equal,omitempty"`
	Contains       string `json:"contains,omitempty"`
	DoesNotContain string `json:"does_not_contain,omitempty"`
	StartsWith     string `json:"starts_with,omitempty"`
	EndsWith       string `json:"ends_with,omitempty"`
	IsEmpty        bool   `json:"is_empty,omitempty"`
	IsNotEmpty     bool   `json:"is_not_empty,omitempty"`
}

type TextProperty ¶

type TextProperty struct {
	ID   PropertyID   `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	Text []RichText   `json:"text"`
}

func (TextProperty) GetID ¶ added in v1.12.9

func (p TextProperty) GetID() string

func (TextProperty) GetType ¶

func (p TextProperty) GetType() PropertyType

type TimestampFilter ¶ added in v1.8.5

type TimestampFilter struct {
	Timestamp      TimestampType        `json:"timestamp"`
	CreatedTime    *DateFilterCondition `json:"created_time,omitempty"`
	LastEditedTime *DateFilterCondition `json:"last_edited_time,omitempty"`
}

type TimestampType ¶

type TimestampType string

const (
	TimestampCreated    TimestampType = "created_time"
	TimestampLastEdited TimestampType = "last_edited_time"
)

type TitleProperty ¶ added in v1.2.0

type TitleProperty struct {
	ID    PropertyID   `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Title []RichText   `json:"title"`
}

func (TitleProperty) GetID ¶ added in v1.12.9

func (p TitleProperty) GetID() string

func (TitleProperty) GetType ¶ added in v1.2.0

func (p TitleProperty) GetType() PropertyType

type TitlePropertyConfig ¶ added in v1.2.0

type TitlePropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Title struct{}           `json:"title"`
}

func (TitlePropertyConfig) GetID ¶ added in v1.13.0

func (p TitlePropertyConfig) GetID() PropertyID

func (TitlePropertyConfig) GetType ¶ added in v1.2.0

func (p TitlePropertyConfig) GetType() PropertyConfigType

type ToDo ¶ added in v1.4.0

type ToDo struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Checked  bool       `json:"checked"`
	Color    string     `json:"color,omitempty"`
}

type ToDoBlock ¶

type ToDoBlock struct {
	BasicBlock
	ToDo ToDo `json:"to_do"`
}

func (ToDoBlock) GetRichTextString ¶ added in v1.12.10

func (t ToDoBlock) GetRichTextString() string

type Toggle ¶

type Toggle struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type ToggleBlock ¶

type ToggleBlock struct {
	BasicBlock
	Toggle Toggle `json:"toggle"`
}

func (ToggleBlock) GetRichTextString ¶ added in v1.12.10

func (b ToggleBlock) GetRichTextString() string

type Token ¶

type Token string

func (Token) String ¶

func (it Token) String() string

type TokenCreateError ¶ added in v1.12.7

type TokenCreateError struct {
	Code    ErrorCode `json:"error"`
	Message string    `json:"error_description"`
}

func (*TokenCreateError) Error ¶ added in v1.12.7

func (e *TokenCreateError) Error() string

type TokenCreateRequest ¶ added in v1.12.0

type TokenCreateRequest struct {
	// A unique random code that Notion generates to authenticate with your service,
	// generated when a user initiates the OAuth flow.
	Code string `json:"code"`
	// A constant string: "authorization_code".
	GrantType string `json:"grant_type"`
	// The "redirect_uri" that was provided in the OAuth Domain & URI section of
	// the integration's Authorization settings. Do not include this field if a
	// "redirect_uri" query param was not included in the Authorization URL
	// provided to users. In most cases, this field is required.
	RedirectUri string `json:"redirect_uri,omitempty"`
	// Required if and only when building Link Preview integrations (otherwise
	// ignored). An object with key and name properties. key should be a unique
	// identifier for the account. Notion uses the key to determine whether or not
	// the user is re-connecting the same account. name should be some way for the
	// user to know which account they used to authenticate with your service. If
	// a user has authenticated Notion with your integration before and key is the
	// same but name is different, then Notion updates the name associated with
	// your integration.
	ExternalAccount ExternalAccount `json:"external_account,omitempty"`
}

TokenCreateRequest represents the request body for AuthenticationClient.CreateToken.
type TokenCreateResponse ¶ added in v1.12.0

type TokenCreateResponse struct {
	AccessToken          string `json:"access_token"`
	BotId                string `json:"bot_id"`
	DuplicatedTemplateId string `json:"duplicated_template_id,omitempty"`

	// Owner can be { "workspace": true } OR a User object.
	// Ref: https://developers.notion.com/docs/authorization#step-4-notion-responds-with-an-access_token-and-some-additional-information
	Owner         interface{} `json:"owner,omitempty"`
	WorkspaceIcon string      `json:"workspace_icon"`
	WorkspaceId   string      `json:"workspace_id"`
	WorkspaceName string      `json:"workspace_name"`
}

type URLProperty ¶

type URLProperty struct {
	ID   ObjectID     `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	URL  string       `json:"url"`
}

func (URLProperty) GetID ¶ added in v1.12.9

func (p URLProperty) GetID() string

func (URLProperty) GetType ¶

func (p URLProperty) GetType() PropertyType

type URLPropertyConfig ¶ added in v1.2.0

type URLPropertyConfig struct {
	ID   PropertyID         `json:"id,omitempty"`
	Type PropertyConfigType `json:"type"`
	URL  struct{}           `json:"url"`
}

func (URLPropertyConfig) GetID ¶ added in v1.13.0

func (p URLPropertyConfig) GetID() PropertyID

func (URLPropertyConfig) GetType ¶ added in v1.2.0

func (p URLPropertyConfig) GetType() PropertyConfigType

type UniqueID ¶ added in v1.12.2

type UniqueID struct {
	Prefix *string `json:"prefix,omitempty"`
	Number int     `json:"number"`
}

func (UniqueID) String ¶ added in v1.12.2

func (uID UniqueID) String() string

type UniqueIDConfig ¶ added in v1.12.9

type UniqueIDConfig struct {
	Prefix string `json:"prefix"`
}

type UniqueIDProperty ¶ added in v1.12.2

type UniqueIDProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	UniqueID UniqueID     `json:"unique_id"`
}

func (UniqueIDProperty) GetID ¶ added in v1.12.9

func (p UniqueIDProperty) GetID() string

func (UniqueIDProperty) GetType ¶ added in v1.12.2

func (p UniqueIDProperty) GetType() PropertyType

type UniqueIDPropertyConfig ¶ added in v1.12.2

type UniqueIDPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	UniqueID UniqueIDConfig     `json:"unique_id"`
}

func (UniqueIDPropertyConfig) GetID ¶ added in v1.13.0

func (p UniqueIDPropertyConfig) GetID() PropertyID

func (UniqueIDPropertyConfig) GetType ¶ added in v1.12.2

func (p UniqueIDPropertyConfig) GetType() PropertyConfigType

type UniqueIdFilterCondition ¶ added in v1.13.1

type UniqueIdFilterCondition struct {
	Equals               *int `json:"equals,omitempty"`
	DoesNotEqual         *int `json:"does_not_equal,omitempty"`
	GreaterThan          *int `json:"greater_than,omitempty"`
	LessThan             *int `json:"less_than,omitempty"`
	GreaterThanOrEqualTo *int `json:"greater_than_or_equal_to,omitempty"`
	LessThanOrEqualTo    *int `json:"less_than_or_equal_to,omitempty"`
}

type UnsupportedBlock ¶ added in v1.5.2

type UnsupportedBlock struct {
	BasicBlock
}

type User ¶

type User struct {
	Object    ObjectType `json:"object,omitempty"`
	ID        UserID     `json:"id"`
	Type      UserType   `json:"type,omitempty"`
	Name      string     `json:"name,omitempty"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	Person    *Person    `json:"person,omitempty"`
	Bot       *Bot       `json:"bot,omitempty"`
}

type UserClient ¶

type UserClient struct {
	// contains filtered or unexported fields
}

func (*UserClient) Get ¶

func (uc *UserClient) Get(ctx context.Context, id UserID) (*User, error)

Retrieves a User using the ID specified.

See https://developers.notion.com/reference/get-user
func (*UserClient) List ¶

func (uc *UserClient) List(ctx context.Context, pagination *Pagination) (*UsersListResponse, error)

Returns a paginated list of Users for the workspace. The response may contain fewer than page_size of results.

See https://developers.notion.com/reference/get-users
func (*UserClient) Me ¶ added in v1.10.2

func (uc *UserClient) Me(ctx context.Context) (*User, error)

Retrieves the bot User associated with the API token provided in the authorization header. The bot will have an owner field with information about the person who authorized the integration.

See https://developers.notion.com/reference/get-self
type UserID ¶

type UserID string

func (UserID) String ¶

func (uID UserID) String() string

type UserService ¶

type UserService interface {
	List(context.Context, *Pagination) (*UsersListResponse, error)
	Get(context.Context, UserID) (*User, error)
	Me(context.Context) (*User, error)
}

type UserType ¶

type UserType string

const (
	UserTypePerson UserType = "person"
	UserTypeBot    UserType = "bot"
)

type UsersListResponse ¶

type UsersListResponse struct {
	Object     ObjectType `json:"object"`
	Results    []User     `json:"results"`
	HasMore    bool       `json:"has_more"`
	NextCursor Cursor     `json:"next_cursor"`
}

type Verification ¶ added in v1.12.9

type Verification struct {
	State      VerificationState `json:"state"`
	VerifiedBy *User             `json:"verified_by,omitempty"`
	Date       *DateObject       `json:"date,omitempty"`
}

Verification documented here: https://developers.notion.com/reference/page-property-values#verification
type VerificationProperty ¶ added in v1.12.9

type VerificationProperty struct {
	ID           ObjectID     `json:"id,omitempty"`
	Type         PropertyType `json:"type,omitempty"`
	Verification Verification `json:"verification"`
}

func (VerificationProperty) GetID ¶ added in v1.12.9

func (p VerificationProperty) GetID() string

func (VerificationProperty) GetType ¶ added in v1.12.9

func (p VerificationProperty) GetType() PropertyType

type VerificationPropertyConfig ¶ added in v1.12.9

type VerificationPropertyConfig struct {
	ID           PropertyID         `json:"id,omitempty"`
	Type         PropertyConfigType `json:"type,omitempty"`
	Verification Verification       `json:"verification"`
}

func (VerificationPropertyConfig) GetID ¶ added in v1.13.0

func (p VerificationPropertyConfig) GetID() PropertyID

func (VerificationPropertyConfig) GetType ¶ added in v1.12.9

func (p VerificationPropertyConfig) GetType() PropertyConfigType

type VerificationState ¶ added in v1.12.9

type VerificationState string

const (
	VerificationStateVerified   VerificationState = "verified"
	VerificationStateUnverified VerificationState = "unverified"
)

func (VerificationState) String ¶ added in v1.12.9

func (vs VerificationState) String() string

type Video ¶ added in v1.5.0

type Video struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type VideoBlock ¶ added in v1.5.0

type VideoBlock struct {
	BasicBlock
	Video Video `json:"video"`
}

func (VideoBlock) GetRichTextString ¶ added in v1.12.10

func (b VideoBlock) GetRichTextString() string

Source Files ¶
View all Source files
    authentication.go
    block.go
    client.go
    comment.go
    const.go
    database.go
    downloadable_interface.go
    error.go
    filter.go
    object.go
    page.go
    property.go
    property_config.go
    search.go
    sort.go
    user.go 
