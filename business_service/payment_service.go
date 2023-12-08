package business_service

import (
	"fmt"
	"log"
	"strings"

	"github.com/zapscloud/golib-business-repository/business_common"
	"github.com/zapscloud/golib-business-repository/business_repository"
	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// PaymentService - Business Payment Service structure
type PaymentService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(paymentId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(paymentId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(paymentId string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// PaymentService - Business Payment Service structure
type paymentBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoPayment  business_repository.PaymentDao
	daoBusiness platform_repository.BusinessDao
	child       PaymentService
	businessId  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewPaymentService(props utils.Map) (PaymentService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("PaymentService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := paymentBaseService{}
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
	p.businessId = businessId
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

// paymentBaseService - Close all the services
func (p *paymentBaseService) EndService() {
	log.Printf("EndPaymentService ")
	p.CloseDatabaseService()
}

func (p *paymentBaseService) initializeService() {
	log.Printf("PaymentMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoPayment = business_repository.NewPaymentDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *paymentBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("PaymentService::FindAll - Begin")

	listdata, err := p.daoPayment.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("PaymentService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *paymentBaseService) Get(paymentId string) (utils.Map, error) {
	log.Printf("PaymentService::Get::  Begin %v", paymentId)

	data, err := p.daoPayment.Get(paymentId)

	log.Println("PaymentService::Get:: End ", err)
	return data, err
}

func (p *paymentBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("PaymentBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoPayment.Find(filter)
	log.Println("PaymentBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *paymentBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("PaymentService::Create - Begin")
	var paymentId string

	dataval, dataok := indata[business_common.FLD_PAYMENT_ID]
	if dataok {
		paymentId = strings.ToLower(dataval.(string))
	} else {
		paymentId = utils.GenerateUniqueId("pay")
		log.Println("Unique Payment ID", paymentId)
	}

	//BusinessPayment
	indata[business_common.FLD_BUSINESS_ID] = p.businessId
	indata[business_common.FLD_PAYMENT_ID] = paymentId

	data, err := p.daoPayment.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("PaymentService::Create - End")
	return data, nil
}

// Update - Update Service
func (p *paymentBaseService) Update(paymentId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessPaymentService::Update - Begin")

	data, err := p.daoPayment.Update(paymentId, indata)

	log.Println("PaymentService::Update - End")
	return data, err
}

// Delete - Delete Service
func (p *paymentBaseService) Delete(paymentId string, delete_permanent bool) error {

	log.Println("PaymentService::Delete - Begin", paymentId)

	if delete_permanent {
		result, err := p.daoPayment.Delete(paymentId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(paymentId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("PaymentService::Delete - End")
	return nil
}

func (p *paymentBaseService) errorReturn(err error) (PaymentService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
