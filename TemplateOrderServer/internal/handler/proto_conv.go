package handler

import pb "template-mall/TemplateOrderServer/api/proto"

func isMemberFromDB(v int8) bool {
	return v == 1
}

func isMemberToDB(v bool) int8 {
	if v {
		return 1
	}
	return 0
}

func isFreeFromDB(v int8) bool {
	return v == 1
}

func isFreeToDB(v bool) int8 {
	if v {
		return 1
	}
	return 0
}

func templateStatusFromDB(v int8) pb.TemplateStatus {
	if v == 1 {
		return pb.TemplateStatus_TEMPLATE_STATUS_ONLINE
	}
	return pb.TemplateStatus_TEMPLATE_STATUS_OFFLINE
}

func templateStatusToDB(s pb.TemplateStatus) (int8, bool) {
	switch s {
	case pb.TemplateStatus_TEMPLATE_STATUS_OFFLINE:
		return 0, true
	case pb.TemplateStatus_TEMPLATE_STATUS_ONLINE:
		return 1, true
	default:
		return 0, false
	}
}

func orderStatusFromDB(v int8) pb.OrderStatus {
	return pb.OrderStatus(v)
}

func orderTypeFromDB(v int8) pb.OrderType {
	return pb.OrderType(v)
}
