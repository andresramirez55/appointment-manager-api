package repositories

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type patientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(db *gorm.DB) *patientRepository {
	return &patientRepository{db: db}
}

func (r *patientRepository) Create(ctx context.Context, patient *models.Patient) error {
	return r.db.WithContext(ctx).Create(patient).Error
}

func (r *patientRepository) Update(ctx context.Context, patient *models.Patient) error {
	return r.db.WithContext(ctx).Save(patient).Error
}

func (r *patientRepository) FindByPhone(ctx context.Context, phone string, professionalID int64) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.WithContext(ctx).Where("phone = ? AND professional_id = ?", phone, professionalID).First(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindByID(ctx context.Context, id int64) (*models.Patient, error) {
	var patient models.Patient
	if err := r.db.WithContext(ctx).First(&patient, id).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (r *patientRepository) FindAll(ctx context.Context, professionalID int64) ([]*models.Patient, error) {
	var patients []*models.Patient
	if err := r.db.WithContext(ctx).Where("professional_id = ?", professionalID).Order("name ASC").Find(&patients).Error; err != nil {
		return nil, err
	}
	return patients, nil
}
