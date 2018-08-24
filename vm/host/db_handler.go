package host

import (
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
)

// DBHandler struct {
type DBHandler struct {
	db  *database.Visitor
	ctx *Context
}

// NewDBHandler ...
func NewDBHandler(db *database.Visitor, ctx *Context) DBHandler {
	return DBHandler{
		db:  db,
		ctx: ctx,
	}
}

// Put ...
func (h *DBHandler) Put(key string, value interface{}) *contract.Cost {
	h.db.Put(
		h.modifyKey(key),
		database.MustMarshal(value),
	)
	return PutCost
}

// Get ...
func (h *DBHandler) Get(key string) (value interface{}, cost *contract.Cost) {

	rtn := database.MustUnmarshal(h.db.Get(h.modifyKey(key)))

	return rtn, GetCost
}

// Del ...
func (h *DBHandler) Del(key string) *contract.Cost {

	h.db.Del(h.modifyKey(key))

	return DelCost
}

// MapPut ...
func (h *DBHandler) MapPut(key, field string, value interface{}) *contract.Cost {
	v := database.MustMarshal(value)
	h.db.MPut(h.modifyKey(key), field, v)
	return PutCost
}

// MapGet ...
func (h *DBHandler) MapGet(key, field string) (value interface{}, cost *contract.Cost) {
	rtn := database.MustUnmarshal(h.db.MGet(h.modifyKey(key), field))
	return rtn, GetCost
}

// MapKeys ...
func (h *DBHandler) MapKeys(key string) (fields []string, cost *contract.Cost) {

	return h.db.MKeys(h.modifyKey(key)), KeysCost
}

// MapDel ...
func (h *DBHandler) MapDel(key, field string) *contract.Cost {
	h.db.MDel(h.modifyKey(key), field)
	return DelCost
}

// MapLen ...
func (h *DBHandler) MapLen(key string) (int, *contract.Cost) {
	return len(h.db.MKeys(h.modifyKey(key))), KeysCost
}

// GlobalGet ...
func (h *DBHandler) GlobalGet(con, key string) (value interface{}, cost *contract.Cost) {
	o := h.db.Get(con + database.Separator + key)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapGet ...
func (h *DBHandler) GlobalMapGet(con, key, field string) (value interface{}, cost *contract.Cost) {
	o := h.db.MGet(con+database.Separator+key, field)
	return database.MustUnmarshal(o), GetCost
}

// GlobalMapKeys ...
func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost *contract.Cost) {
	return h.db.MKeys(con + database.Separator + key), GetCost
}

// GlobalMapLen ...
func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost *contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string) string {
	return h.ctx.Value("contract_name").(string) + database.Separator + key
}