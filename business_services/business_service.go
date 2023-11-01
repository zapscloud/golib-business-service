package business_services

import (
	"fmt"
	"log"

	"github.com/zapscloud/golib-business/business_common"
	"github.com/zapscloud/golib-business/business_repository"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-utils/utils"
)

// BusinessService - Users Service structure
type BusinessService interface {
	GetDetails() (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(indata utils.Map) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Delete() error

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()
	EndService()
}

// BusinessService - Users Service structure
type businessBaseService struct {
	db_utils.DatabaseService
	dbRegion            db_utils.DatabaseService
	daoBusiness         business_repository.BusinessDao
	daoUser             business_repository.UserDao
	daoContact          business_repository.ContactDao
	daoSysUser          platform_repository.SysUserDao
	daoPlatformBusiness platform_repository.BusinessDao
	child               BusinessService
	businessID          string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

// NewBusinessService - Construct Business Service
func NewBusinessService(props utils.Map) (BusinessService, error) {
	funcode := business_common.GetServiceModuleCode() + "M" + "01"
	log.Printf("BusinessService :: Start")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, business_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := businessBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		log.Println("NewBusinessMongoService App Connection Error ", err)
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

	// Initialise Services
	p.initializeService()

	_, err = p.daoPlatformBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{ErrorCode: funcode + "01", ErrorMsg: "Invalid business_id", ErrorDetail: "Given business_id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// EndService - Close all the services
func (p *businessBaseService) EndService() {
	log.Printf("EndBusinessMongoService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *businessBaseService) initializeService() {
	log.Printf("BusinessMongoService:: GetBusinessDao ")
	p.daoSysUser = platform_repository.NewSysUserDao(p.GetClient())
	p.daoPlatformBusiness = platform_repository.NewBusinessDao(p.GetClient())

	p.daoBusiness = business_repository.NewBusinessDao(p.dbRegion.GetClient(), p.businessID)
	p.daoUser = business_repository.NewUserDao(p.dbRegion.GetClient(), p.businessID)
	p.daoContact = business_repository.NewContactDao(p.dbRegion.GetClient(), p.businessID)
}

// Create - Create Service
func (p *businessBaseService) Create(indata utils.Map) (utils.Map, error) {

	funcode := business_common.GetServiceModuleCode() + "01"

	log.Println("BusinessService::Create - Begin")

	// Add Business Id
	indata[business_common.FLD_BUSINESS_ID] = p.businessID

	// Create Business
	dataBusiness, err := p.daoBusiness.Create(indata)
	if err != nil {
		log.Println("Business Create Error  ", err)
		err := &utils.AppError{ErrorCode: funcode + "02", ErrorMsg: "Business Create Error", ErrorDetail: "Error while creating business tenant"}
		return nil, err
	}

	log.Println("Business create  ", dataBusiness)

	log.Println("BusinessService::Create - End ")
	return dataBusiness, nil
}

// Update - Update Service
func (p *businessBaseService) Update(indata utils.Map) (utils.Map, error) {

	log.Println("BusinessService::Update - Begin")

	data, err := p.daoBusiness.Get(p.businessID)
	if err != nil {
		return data, err
	}

	// Delete Business Id if exist
	delete(indata, business_common.FLD_BUSINESS_ID)

	data, err = p.daoBusiness.Update(indata)
	log.Println("BusinessService::Update - End ")
	return data, err
}

// FindByCode - Find By Code
func (p *businessBaseService) GetDetails() (utils.Map, error) {
	funcode := business_common.GetServiceModuleCode() + "02"
	log.Printf("BusinessService::GetDetails::  Begin %v", p.businessID)

	data, err := p.daoBusiness.Get(p.businessID)
	if err != nil {
		err := &utils.AppError{ErrorCode: funcode + "02", ErrorMsg: "Invalid businessId", ErrorDetail: "Given businessId is not exist"}
		return nil, err
	}

	log.Println("BusinessService::GetDetails:: End ", err)
	return data, err
}

func (p *businessBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("BusinessService::FindByCode::  Begin ", filter)

	data, err := p.daoContact.Find(filter)
	log.Println("BusinessService::FindByCode:: End ", data, err)
	return data, err
}

// Delete - Delete Service
func (p *businessBaseService) Delete() error {

	log.Println("BusinessService::Delete - Begin", p.businessID)

	result, err := p.daoContact.DeleteAll()
	if err != nil {
		return err
	}
	log.Printf("BusinessService::DeleteAll - Contact  %v", result)

	result, err = p.daoBusiness.Delete(p.businessID)
	if err != nil {
		return err
	}
	log.Printf("BusinessService::Delete - End %v", result)
	return nil
}

func (p *businessBaseService) errorReturn(err error) (BusinessService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
