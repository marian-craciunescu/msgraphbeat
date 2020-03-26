package beater

type OData struct {
	ODataContext  string `json:"@odata.context,omitempty"`
	ODataCount    int    `json:"@odata.count,omitempty"`
	ODataNextLink string `json:"@odata.nextLink,omitempty"`
	ODataETag     string `json:"@odata.etag,omitempty"`
}

type AlertResponse struct {
	OData
	Value []map[string]interface{} `json:"value"`
}
