package errors

// 业务错误码
const (
	// 通用
	ErrSuccess      = 0
	ErrServer       = 10000
	ErrParam        = 10001 // 错误参数
	ErrUnauthorized = 10002
	ErrForbidden    = 10003
	ErrNotFound     = 10004
	ErrConflict     = 10005 // 资源冲突（幂等冲突、唯一键冲突等）
	ErrTooManyReq   = 10006 // 频率限制
	ErrServiceBusy  = 10007 // 服务繁忙（如队列满、系统降级）

	// 用户模块 11xxx
	ErrUserNotFound      = 11001
	ErrUserAlreadyExists = 11002
	ErrPasswordWrong     = 11003
	ErrTokenInvalid      = 11004
	ErrTokenExpired      = 11005
	ErrUserDisabled      = 11006

	// 商品模块 12xxx
	ErrProductNotFound   = 12001
	ErrProductOutOfStock = 12002
	ErrProductDisabled   = 12003

	// 订单模块 13xxx
	ErrOrderNotFound      = 13001
	ErrOrderStatusInvalid = 13002
	ErrOrderDuplicate     = 13003

	// 支付模块 14xxx
	ErrPaymentFailed         = 14001
	ErrPaymentDuplicate      = 14002
	ErrPaymentNotFound       = 14003
	ErrPaymentStatusInvalid  = 14004
	ErrPaymentChannelInvalid = 14005

	// 秒杀模块 15xxx
	ErrSeckillOver       = 15001
	ErrSeckillRepeat     = 15002
	ErrSeckillNotStarted = 15003
	ErrSeckillDisabled   = 15004
	ErrSeckillQueueFull  = 15005

	// 购物车模块 16xxx
	ErrCartEmpty        = 16001
	ErrCartItemNotFound = 16002
	ErrCartItemInvalid  = 16003
	ErrCartItemConflict = 16004
)

var msgMap = map[int]string{
	ErrSuccess:               "success",
	ErrServer:                "服务器内部错误",
	ErrParam:                 "参数错误",
	ErrUnauthorized:          "未授权",
	ErrForbidden:             "无权限",
	ErrNotFound:              "资源不存在",
	ErrConflict:              "资源冲突",
	ErrTooManyReq:            "请求过于频繁",
	ErrServiceBusy:           "服务繁忙，请稍后重试",
	ErrUserNotFound:          "用户不存在",
	ErrUserAlreadyExists:     "用户已存在",
	ErrPasswordWrong:         "密码错误",
	ErrTokenInvalid:          "token无效",
	ErrTokenExpired:          "token已过期",
	ErrUserDisabled:          "用户已禁用",
	ErrProductNotFound:       "商品不存在",
	ErrProductOutOfStock:     "商品库存不足",
	ErrProductDisabled:       "商品已下架",
	ErrOrderNotFound:         "订单不存在",
	ErrOrderStatusInvalid:    "订单状态异常",
	ErrOrderDuplicate:        "重复下单",
	ErrPaymentFailed:         "支付失败",
	ErrPaymentDuplicate:      "重复支付",
	ErrPaymentNotFound:       "支付记录不存在",
	ErrPaymentStatusInvalid:  "支付状态异常",
	ErrPaymentChannelInvalid: "支付渠道不支持",
	ErrSeckillOver:           "秒杀已结束",
	ErrSeckillRepeat:         "请勿重复秒杀",
	ErrSeckillNotStarted:     "秒杀未开始",
	ErrSeckillDisabled:       "秒杀活动已下架",
	ErrSeckillQueueFull:      "秒杀队列繁忙，请重试",
	ErrCartEmpty:             "购物车为空",
	ErrCartItemNotFound:      "购物车商品不存在",
	ErrCartItemInvalid:       "购物车商品无效",
	ErrCartItemConflict:      "购物车商品冲突",
}

// 根据错误码返回错误描述
func Msg(code int) string {
	if msg, ok := msgMap[code]; ok {
		return msg
	}
	return "未知错误"
}
