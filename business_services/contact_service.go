package business_services

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

// ContactService - Contacts Service structure
type ContactService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	GetDetails(contact_id string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(contact_id string, indata utils.Map) (utils.Map, error)
	Delete(contact_id string) error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// ContactBaseService - Contacts Service structure
type contactBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoContact  business_repository.ContactDao
	daoBusiness platform_repository.BusinessDao
	child       ContactService
	businessID  string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewContactService(props utils.Map) (ContactService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("ContactMongoService::Start")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := contactBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}
	// Open RegionDB Service
	p.dbRegion, err = platform_services.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessID = businessId

	// Initialize other Service
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{ErrorCode: funcode + "01", ErrorMsg: "Invalid business_id", ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// ContactBaseService - Close all the services
func (p *contactBaseService) EndService() {
	log.Printf("EndContactMongoService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *contactBaseService) initializeService() {
	log.Printf("ContactMongoService:: GetBusinessDao ")
	p.daoContact = business_repository.NewContactDao(p.dbRegion.GetClient(), p.businessID)
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
}

// List - List All records
func (p *contactBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("ContactService::FindAll - Begin")

	daoContact := p.daoContact
	response, err := daoContact.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("ContactService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *contactBaseService) GetDetails(contact_id string) (utils.Map, error) {
	log.Printf("ContactService::FindByCode::  Begin %v", contact_id)

	data, err := p.daoContact.Get(contact_id)
	log.Println("ContactService::FindByCode:: End ", err)
	return data, err
}

func (p *contactBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("ContactService::FindByCode::  Begin ", filter)

	data, err := p.daoContact.Find(filter)
	log.Println("ContactService::FindByCode:: End ", data, err)
	return data, err
}

func (p *contactBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	dataval, dataok := indata[business_common.FLD_APP_CONTACT_ID]
	if !dataok {
		uid := utils.GenerateUniqueId("cont")
		log.Println("Unique Contact ID", uid)
		indata[business_common.FLD_APP_CONTACT_ID] = uid
		dataval = indata[business_common.FLD_APP_CONTACT_ID]
	}
	indata[business_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided Contact ID:", dataval)

	_, err := p.daoContact.Get(dataval.(string))
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing Contact ID !", ErrorDetail: "Given Contact ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoContact.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *contactBaseService) Update(contact_id string, indata utils.Map) (utils.Map, error) {

	log.Println("ContactService::Update - Begin")

	data, err := p.daoContact.Get(contact_id)
	if err != nil {
		return data, err
	}

	data, err = p.daoContact.Update(contact_id, indata)
	log.Println("ContactService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *contactBaseService) Delete(contact_id string) error {

	log.Println("ContactService::Delete - Begin", contact_id)

	daoContact := p.daoContact
	result, err := daoContact.Delete(contact_id)
	if err != nil {
		return err
	}

	log.Printf("ContactService::Delete - End %v", result)
	return nil
}

func (p *contactBaseService) errorReturn(err error) (ContactService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
