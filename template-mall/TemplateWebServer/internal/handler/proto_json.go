package handler

import pb "template-mall/TemplateOrderServer/api/proto"

func memberStatusJSON(isMember bool) int32 {
	if isMember {
		return 1
	}
	return 0
}

func isFreeJSON(isFree bool) int32 {
	if isFree {
		return 1
	}
	return 0
}

func templateStatusJSON(s pb.TemplateStatus) int32 {
	if s == pb.TemplateStatus_TEMPLATE_STATUS_ONLINE {
		return 1
	}
	return 0
}

func orderStatusJSON(s pb.OrderStatus) int32 {
	return int32(s)
}

func orderTypeJSON(t pb.OrderType) int32 {
	return int32(t)
}

func templateStatusFromQuery(status int) *pb.TemplateStatus {
	switch status {
	case 0:
		s := pb.TemplateStatus_TEMPLATE_STATUS_OFFLINE
		return &s
	case 1:
		s := pb.TemplateStatus_TEMPLATE_STATUS_ONLINE
		return &s
	default:
		return nil
	}
}

func orderStatusFromQuery(status int) *pb.OrderStatus {
	if status < 1 || status > 3 {
		return nil
	}
	s := pb.OrderStatus(status)
	return &s
}

func templateStatusFromJSON(status int32) pb.TemplateStatus {
	if status == 1 {
		return pb.TemplateStatus_TEMPLATE_STATUS_ONLINE
	}
	return pb.TemplateStatus_TEMPLATE_STATUS_OFFLINE
}
