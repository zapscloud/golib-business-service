package business_service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/zapscloud/golib-business-repository/business_common"
	"github.com/zapscloud/golib-business-repository/business_repository"
	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// PaymentTxnService - Business PaymentTxn Service structure
type PaymentTxnService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(PaymentTxnId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(PaymentTxnId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(PaymentTxnId string, delete_permanent bool) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// PaymentTxnService - Business PaymentTxn Service structure
type PaymentTxnBaseService struct {
	db_utils.DatabaseService
	dbRegion      db_utils.DatabaseService
	daoPaymentTxn business_repository.PaymentTxnDao
	daoBusiness   platform_repository.BusinessDao
	child         PaymentTxnService
	businessId    string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewPaymentTxnService(props utils.Map) (PaymentTxnService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("PaymentTxnService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := PaymentTxnBaseService{}
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

// PaymentTxnBaseService - Close all the services
func (p *PaymentTxnBaseService) EndService() {
	log.Printf("EndPaymentTxnService ")
	p.CloseDatabaseService()
}

func (p *PaymentTxnBaseService) initializeService() {
	log.Printf("PaymentTxnMongoService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoPaymentTxn = business_repository.NewPaymentTxnDao(p.dbRegion.GetClient(), p.businessId)
}

// List - List All records
func (p *PaymentTxnBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("PaymentTxnService::FindAll - Begin")

	listdata, err := p.daoPaymentTxn.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("PaymentTxnService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *PaymentTxnBaseService) Get(PaymentTxnId string) (utils.Map, error) {
	log.Printf("PaymentTxnService::Get::  Begin %v", PaymentTxnId)

	data, err := p.daoPaymentTxn.Get(PaymentTxnId)

	log.Println("PaymentTxnService::Get:: End ", err)
	return data, err
}

func (p *PaymentTxnBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("PaymentTxnBaseService::FindByCode::  Begin ", filter)

	data, err := p.daoPaymentTxn.Find(filter)
	log.Println("PaymentTxnBaseService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *PaymentTxnBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("PaymentTxnService::Create - Begin")
	var PaymentTxnId string

	dataval, dataok := indata[business_common.FLD_PAYMENT_TXN_ID]
	if dataok {
		PaymentTxnId = strings.ToLower(dataval.(string))
	} else {
		PaymentTxnId = utils.GenerateUniqueId("paytxn")
		log.Println("Unique PaymentTxn ID", PaymentTxnId)
	}
	dateTime := time.Now().Format(time.DateTime)
	//BusinessPaymentTxn
	indata[business_common.FLD_DATE_TIME] = dateTime
	indata[business_common.FLD_BUSINESS_ID] = p.businessId
	indata[business_common.FLD_PAYMENT_TXN_ID] = PaymentTxnId


	data, err := p.daoPaymentTxn.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("PaymentTxnService::Create - End")
	return data, nil
}

// Update - Update Service
func (p *PaymentTxnBaseService) Update(PaymentTxnId string, indata utils.Map) (utils.Map, error) {

	log.Println("BusinessPaymentTxnService::Update - Begin")

	data, err := p.daoPaymentTxn.Update(PaymentTxnId, indata)

	log.Println("PaymentTxnService::Update - End")
	return data, err
}

// Delete - Delete Service
func (p *PaymentTxnBaseService) Delete(PaymentTxnId string, delete_permanent bool) error {

	log.Println("PaymentTxnService::Delete - Begin", PaymentTxnId)

	if delete_permanent {
		result, err := p.daoPaymentTxn.Delete(PaymentTxnId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(PaymentTxnId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("PaymentTxnService::Delete - End")
	return nil
}

func (p *PaymentTxnBaseService) errorReturn(err error) (PaymentTxnService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
