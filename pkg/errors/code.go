package errors

// 业务错误码
const (
	// 通用
	ErrSuccess      = 0
	ErrServer       = 10000
	ErrParam        = 10001			// 错误参数
	ErrUnauthorized = 10002
	ErrForbidden    = 10003
	ErrNotFound     = 10004

	// 用户模块 11xxx
	ErrUserNotFound      = 11001
	ErrUserAlreadyExists = 11002
	ErrPasswordWrong     = 11003
	ErrTokenInvalid      = 11004
	ErrTokenExpired      = 11005

	// 商品模块 12xxx
	ErrProductNotFound    = 12001
	ErrProductOutOfStock  = 12002

	// 订单模块 13xxx
	ErrOrderNotFound      = 13001
	ErrOrderStatusInvalid = 13002

	// 支付模块 14xxx
	ErrPaymentFailed      = 14001
	ErrPaymentDuplicate   = 14002

	// 秒杀模块 15xxx
	ErrSeckillOver        = 15001
	ErrSeckillRepeat      = 15002
)

var msgMap = map[int]string{
	ErrSuccess:           "success",
	ErrServer:            "服务器内部错误",
	ErrParam:             "参数错误",
	ErrUnauthorized:      "未授权",
	ErrForbidden:         "无权限",
	ErrNotFound:          "资源不存在",
	ErrUserNotFound:      "用户不存在",
	ErrUserAlreadyExists: "用户已存在",
	ErrPasswordWrong:     "密码错误",
	ErrTokenInvalid:      "token无效",
	ErrTokenExpired:      "token已过期",
	ErrProductNotFound:   "商品不存在",
	ErrProductOutOfStock: "商品库存不足",
	ErrOrderNotFound:     "订单不存在",
	ErrOrderStatusInvalid:"订单状态异常",
	ErrPaymentFailed:     "支付失败",
	ErrPaymentDuplicate:  "重复支付",
	ErrSeckillOver:       "秒杀已结束",
	ErrSeckillRepeat:     "请勿重复秒杀",
}

// 根据错误码返回错误描述
func Msg(code int) string {
	if msg, ok := msgMap[code]; ok {
		return msg
	}
	return "未知错误"
}
