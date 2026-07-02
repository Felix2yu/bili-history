package models

type HistoryRecord struct {
	ID           int64  `json:"id" db:"id"`
	Title        string `json:"title" db:"title"`
	LongTitle    string `json:"long_title" db:"long_title"`
	Cover        string `json:"cover" db:"cover"`
	Covers       string `json:"covers" db:"covers"`
	URI          string `json:"uri" db:"uri"`
	OID          int64  `json:"oid" db:"oid"`
	Epid         int64  `json:"epid" db:"epid"`
	Bvid         string `json:"bvid" db:"bvid"`
	Page         int    `json:"page" db:"page"`
	Cid          int64  `json:"cid" db:"cid"`
	Part         string `json:"part" db:"part"`
	Business     string `json:"business" db:"business"`
	Dt           int    `json:"dt" db:"dt"`
	Videos       int    `json:"videos" db:"videos"`
	AuthorName   string `json:"author_name" db:"author_name"`
	AuthorFace   string `json:"author_face" db:"author_face"`
	AuthorMid    int64  `json:"author_mid" db:"author_mid"`
	ViewAt       int64  `json:"view_at" db:"view_at"`
	Progress     int    `json:"progress" db:"progress"`
	Badge        string `json:"badge" db:"badge"`
	ShowTitle    string `json:"show_title" db:"show_title"`
	Duration     int    `json:"duration" db:"duration"`
	Current      string `json:"current" db:"current"`
	Total        int    `json:"total" db:"total"`
	NewDesc      string `json:"new_desc" db:"new_desc"`
	IsFinish     int    `json:"is_finish" db:"is_finish"`
	IsFav        int    `json:"is_fav" db:"is_fav"`
	Kid          int64  `json:"kid" db:"kid"`
	TagName      string `json:"tag_name" db:"tag_name"`
	LiveStatus   int    `json:"live_status" db:"live_status"`
	MainCategory string `json:"main_category" db:"main_category"`
	Remark       string `json:"remark" db:"remark"`
	RemarkTime   int64  `json:"remark_time" db:"remark_time"`
}

type HistoryRecordWithExtra struct {
	HistoryRecord
	OriginalURL string   `json:"original_url"`
	CoversList  []string `json:"covers_list"`
	ViewTime    string   `json:"view_time"`
}

type DeletedHistory struct {
	ID         int64  `json:"id" db:"id"`
	Bvid       string `json:"bvid" db:"bvid"`
	ViewAt     int64  `json:"view_at" db:"view_at"`
	DeleteTime int64  `json:"delete_time" db:"delete_time"`
}

type VideoCategory struct {
	MainCategory string `json:"main_category" db:"main_category"`
	SubCategory  string `json:"sub_category" db:"sub_category"`
	Alias        string `json:"alias" db:"alias"`
	Tid          int    `json:"tid" db:"tid"`
	Image        string `json:"image" db:"image"`
}

type CategoryNode struct {
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	SubCategories []SubCategory `json:"sub_categories"`
}

type SubCategory struct {
	Name  string `json:"name"`
	Alias string `json:"alias"`
	Tid   int    `json:"tid"`
}

type PagedResponse struct {
	Records interface{} `json:"records"`
	Total   int64       `json:"total"`
	Size    int         `json:"size"`
	Current int         `json:"current"`
}

type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse(data interface{}) ApiResponse {
	return ApiResponse{
		Status: "success",
		Data:   data,
	}
}

func ErrorResponse(message string) ApiResponse {
	return ApiResponse{
		Status:  "error",
		Message: message,
	}
}
