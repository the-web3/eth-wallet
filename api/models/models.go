package models

type QueryDWParams struct {
	Address  string
	Page     int
	PageSize int
	Order    string
}

type QueryPageParams struct {
	Page     int
	PageSize int
	Order    string
}

type QueryIdParams struct {
	Id uint64
}

type QueryIndexParams struct {
	Index uint64
}

type DepositsResponse struct {
	Current int   `json:"Current"`
	Size    int   `json:"Size"`
	Total   int64 `json:"Total"`
}

type WithdrawsResponse struct {
	Current int   `json:"Current"`
	Size    int   `json:"Size"`
	Total   int64 `json:"Total"`
}
