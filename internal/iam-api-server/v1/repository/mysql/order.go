package mysql

var allowedOrderBy = map[string]string{
	"id":         "id",
	"created_at": "created_at",
	"updated_at": "updated_at",
	"instanceID": "instanceID",
	"username":   "username",
	"nickname":   "nickname",
	"name":       "name",
	"email":      "email",
	"phone":      "phone",
}
