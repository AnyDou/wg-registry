package store

import (
	"time"

	"github.com/jinzhu/gorm"

	"github.com/sosedoff/wg-registry/model"
	"github.com/sosedoff/wg-registry/util"
)

type DbStore struct {
	db *gorm.DB
}

func NewDatabaseStore(scheme, connstr string) (*DbStore, error) {
	db, err := gorm.Open(scheme, connstr)
	if err != nil {
		return nil, err
	}
	return &DbStore{db: db}, nil
}

func (s *DbStore) AutoMigrate() error {
	return s.db.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.Server{},
	).Error
}

func (s *DbStore) UserCount() (int, error) {
	count := 0
	err := s.db.Model(&model.User{}).Count(&count).Error
	return count, err
}

func (s *DbStore) FindUserByID(id interface{}) (*model.User, error) {
	user := &model.User{}
	err := s.db.First(user, "id = ?", id).Error
	if err == gorm.ErrRecordNotFound {
		user = nil
		err = nil
	}
	return user, err
}

func (s *DbStore) FindUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := s.db.First(user, "LOWER(email) = ?", email).Error
	if err == gorm.ErrRecordNotFound {
		user = nil
		err = nil
	}
	return user, err
}

func (s *DbStore) FindServer() (*model.Server, error) {
	server := &model.Server{}
	err := s.db.First(server).Error
	if err == gorm.ErrRecordNotFound {
		server = nil
		err = nil
	}
	return server, err
}

func (s *DbStore) SaveServer(record *model.Server) error {
	if record.ID == 0 {
		return s.db.Create(record).Error
	}
	return s.db.Update(record).Error
}

func (s *DbStore) CreateServer(server *model.Server) error {
	if err := server.Validate(); err != nil {
		return err
	}
	return s.db.Create(server).Error
}

func (s *DbStore) FindDevice(id interface{}) (*model.Device, error) {
	result := &model.Device{}
	err := s.db.Where("id = ?", id).Find(&result).Error
	return result, err
}

func (s *DbStore) FindUserDevice(user *model.User, id interface{}) (*model.Device, error) {
	result := &model.Device{}
	err := s.db.
		Where("user_id = ? AND id = ?", user.ID, id).
		First(&result).
		Error
	if err == gorm.ErrRecordNotFound {
		result = nil
	}
	return result, err
}

func (s *DbStore) AllDevices() ([]model.Device, error) {
	result := []model.Device{}
	err := s.db.Find(&result).Error
	return result, err
}

func (s *DbStore) AllUsers() ([]model.User, error) {
	result := []model.User{}
	err := s.db.Find(&result).Error
	return result, err
}

func (s *DbStore) GetDevicesByUser(id interface{}) ([]model.Device, error) {
	result := []model.Device{}
	err := s.db.Where("user_id = ?", id).Find(&result).Error
	return result, err
}

func (s *DbStore) CreateUser(user *model.User) error {
	return s.db.Create(user).Error
}

func (s *DbStore) CreateDevice(server *model.Server, device *model.Device) error {
	ip, err := NextIPV4(s, server)
	if err != nil {
		return err
	}
	if err := device.AssignPrivateKey(); err != nil {
		return err
	}

	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	device.IPV4 = ip

	return util.ErrChain(
		func() error { return device.Validate() },
		func() error { return s.db.Create(device).Error },
	)
}

func (s *DbStore) DeleteUserDevice(user *model.User, device *model.Device) error {
	return s.db.Delete(device).Error
}

func (s *DbStore) AllocatedIPV4() ([]string, error) {
	addrs := []string{}
	err := s.db.Model(&model.Device{}).Pluck("ip_v4", &addrs).Error
	return addrs, err
}
