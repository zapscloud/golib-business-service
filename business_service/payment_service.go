package business_service

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business-repository/business_common"
	"github.com/zapscloud/golib-business-repository/business_repository"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// PaymentService - Payments Service structure
type PaymentService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	GetDetails(paymentid string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(paymentid string, indata utils.Map) (utils.Map, error)
	Delete(paymentid string) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// PaymentBaseService - Payments Service structure
type PaymentBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoPayment  business_repository.PaymentDao
	daoBusiness platform_repository.BusinessDao
	child       PaymentService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewPaymentService(props utils.Map) (PaymentService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Print("PaymentService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := PaymentBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_service.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessID = businessId
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid business_id",
			ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// PaymentBaseService - Close all the services
func (p *PaymentBaseService) EndService() {
	log.Printf("EndPaymentService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *PaymentBaseService) initializeService() {
	log.Printf("PaymentService:: GetBusinessDao ")
	p.daoPayment = business_repository.NewPaymentDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *PaymentBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("PaymentService::FindAll - Begin")

	daoPayment := p.daoPayment
	response, err := daoPayment.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("PaymentService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *PaymentBaseService) GetDetails(paymentid string) (utils.Map, error) {
	log.Printf("PaymentService::FindByCode::  Begin %v", paymentid)

	data, err := p.daoPayment.Get(paymentid)
	log.Println("PaymentService::FindByCode:: End ", err)
	return data, err
}

func (p *PaymentBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("PaymentService::FindByCode::  Begin ", filter)

	data, err := p.daoPayment.Find(filter)
	log.Println("PaymentService::FindByCode:: End ", data, err)
	return data, err
}

func (p *PaymentBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[business_common.FLD_PAYMENT_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("payment")
		log.Println("Unique Payment ID", uid)
		indata[business_common.FLD_PAYMENT_ID] = uid
		dataval = indata[business_common.FLD_PAYMENT_ID]
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Payment ID:", dataval)

	_, err := p.daoPayment.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Payment ID !", ErrorDetail: "Given PaymentID already exist"}
		return indata, err
	}

	insertResult, err := p.daoPayment.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *PaymentBaseService) Update(paymentid string, indata utils.Map) (utils.Map, error) {

	log.Println("PaymentService::Update - Begin")

	data, err := p.daoPayment.Get(paymentid)
	if err != nil {
		return data, err
	}

	data, err = p.daoPayment.Update(paymentid, indata)
	log.Println("PaymentService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *PaymentBaseService) Delete(paymentid string) error {

	log.Println("PaymentService::Delete - Begin", paymentid)

	daoPayment := p.daoPayment
	result, err := daoPayment.Delete(paymentid)
	if err != nil {
		return err
	}

	log.Printf("PaymentService::Delete - End %v", result)
	return nil
}

func (p *PaymentBaseService) errorReturn(err error) (PaymentService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
