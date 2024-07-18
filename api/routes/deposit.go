package routes

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
)

func (h Routes) DepositListHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	pageQuery := r.URL.Query().Get("page")
	pageSizeQuery := r.URL.Query().Get("pageSize")
	order := r.URL.Query().Get("order")
	params, err := h.svc.QueryDWListParams(address, pageQuery, pageSizeQuery, order)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		log.Error("error reading request params", "err", err.Error())
		return
	}

	depositPage, err := h.svc.GetDepositList(params)
	if err != nil {
		http.Error(w, "Internal server error reading state root list", http.StatusInternalServerError)
		log.Error("Unable to read state root list from DB", "err", err.Error())
		return
	}

	err = jsonResponse(w, depositPage, http.StatusOK)
	if err != nil {
		log.Error("Error writing response", "err", err.Error())
	}
}
