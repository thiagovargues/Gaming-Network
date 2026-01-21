package apierrors

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
