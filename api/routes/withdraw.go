package routes

import (
	"net/http"

	"github.com/ethereum/go-ethereum/log"
)

func (h Routes) WithdrawListHandler(w http.ResponseWriter, r *http.Request) {
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

	withdrawPage, err := h.svc.GetWithdrawalList(params)
	if err != nil {
		http.Error(w, "Internal server error reading state root list", http.StatusInternalServerError)
		log.Error("Unable to read state root list from DB", "err", err.Error())
		return
	}

	err = jsonResponse(w, withdrawPage, http.StatusOK)
	if err != nil {
		log.Error("Error writing response", "err", err.Error())
	}
}

func (h Routes) SubmitWithdrawHandler(w http.ResponseWriter, r *http.Request) {
	fromAddress := r.URL.Query().Get("fromAddress")
	toaAdress := r.URL.Query().Get("toAddress")
	tokenAddress := r.URL.Query().Get("tokenAddress")
	amount := r.URL.Query().Get("amount")

	params, err := h.svc.SubmitDWParams(fromAddress, toaAdress, tokenAddress, amount)
	if err != nil {
		http.Error(w, "invalid query params", http.StatusBadRequest)
		log.Error("error reading request params", "err", err.Error())
		return
	}
	withdrawRet, err := h.svc.SubmitWithdrawFromBusiness(params)
	if err != nil {
		http.Error(w, "Internal server error reading state root list", http.StatusInternalServerError)
		log.Error("Unable to read state root list from DB", "err", err.Error())
		return
	}
	err = jsonResponse(w, withdrawRet, http.StatusOK)
	if err != nil {
		log.Error("Error writing response", "err", err.Error())
	}
}
