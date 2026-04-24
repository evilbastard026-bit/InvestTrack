package models

import (
	"time"
)

type Owner struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"not null"`
	Company   string    `json:"company"`
	Bio       string    `json:"bio"`
	PhotoURL  string    `json:"photoUrl"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Objects []InvestmentObject `json:"objects,omitempty" gorm:"foreignKey:OwnerID"`
}

type Investor struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"not null"`
	Company   string    `json:"company"`
	Bio       string    `json:"bio"`
	PhotoURL  string    `json:"photoUrl"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Objects []InvestmentObject `json:"objects,omitempty" gorm:"foreignKey:InvestorID"`
}

type ObjectImage struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ObjectID  uint      `json:"objectId" gorm:"not null"`
	ImageURL  string    `json:"imageUrl" gorm:"not null"`
	MediaType string    `json:"type" gorm:"column:media_type;default:'image'"`
	Caption   string    `json:"caption"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type InvestmentObject struct {
    //существующие поля (ID, Title, Description, Location и т.д.) ...
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string    `json:"title" gorm:"not null"`
	Description     string    `json:"description"`
	Type            string    `json:"type" gorm:"not null;default:'building'"`
	Status          string    `json:"status" gorm:"not null;default:'active'"`
	Location        string    `json:"location"`
	ImageURL        string    `json:"imageUrl"`
	InvestmentValue float64   `json:"investmentValue"`
	ReturnRate      float64   `json:"returnRate"`
	Featured        bool      `json:"featured" gorm:"default:false"`
	TopRated        bool      `json:"topRated" gorm:"default:false"`
	TopRating       float64   `json:"topRating" gorm:"column:top_rating;default:0"`
	OwnerID         *uint     `json:"ownerId"`
	InvestorID      *uint     `json:"investorId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`

	Owner         *Owner        `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
	Investor      *Investor     `json:"investor,omitempty" gorm:"foreignKey:InvestorID"`
	Comments      []Comment     `json:"comments,omitempty" gorm:"foreignKey:ObjectID"`
	Images        []ObjectImage `json:"images,omitempty" gorm:"foreignKey:ObjectID"`
	AverageRating *float64      `json:"averageRating" gorm:"-"`
	CommentCount  int           `json:"commentCount" gorm:"-"`

    // Поля для переводов
    TitleRu      string `json:"titleRu" gorm:"column:title_ru"`
    TitleEn      string `json:"titleEn" gorm:"column:title_en"`
    TitleUz      string `json:"titleUz" gorm:"column:title_uz"`
    TitleKk      string `json:"titleKk" gorm:"column:title_kk"`

    DescriptionRu string `json:"descriptionRu" gorm:"column:description_ru"`
    DescriptionEn string `json:"descriptionEn" gorm:"column:description_en"`
    DescriptionUz string `json:"descriptionUz" gorm:"column:description_uz"`
    DescriptionKk string `json:"descriptionKk" gorm:"column:description_kk"`

    LocationRu string `json:"locationRu" gorm:"column:location_ru"`
    LocationEn string `json:"locationEn" gorm:"column:location_en"`
    LocationUz string `json:"locationUz" gorm:"column:location_uz"`
    LocationKk string `json:"locationKk" gorm:"column:location_kk"`

	
}

// type InvestmentObject struct {
// 	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
// 	Title           string    `json:"title" gorm:"not null"`
// 	Description     string    `json:"description"`
// 	Type            string    `json:"type" gorm:"not null;default:'building'"`
// 	Status          string    `json:"status" gorm:"not null;default:'active'"`
// 	Location        string    `json:"location"`
// 	ImageURL        string    `json:"imageUrl"`
// 	InvestmentValue float64   `json:"investmentValue"`
// 	ReturnRate      float64   `json:"returnRate"`
// 	Featured        bool      `json:"featured" gorm:"default:false"`
// 	TopRated        bool      `json:"topRated" gorm:"default:false"`
// 	TopRating       float64   `json:"topRating" gorm:"column:top_rating;default:0"`
// 	OwnerID         *uint     `json:"ownerId"`
// 	InvestorID      *uint     `json:"investorId"`
// 	CreatedAt       time.Time `json:"createdAt"`
// 	UpdatedAt       time.Time `json:"updatedAt"`

// 	Owner         *Owner        `json:"owner,omitempty" gorm:"foreignKey:OwnerID"`
// 	Investor      *Investor     `json:"investor,omitempty" gorm:"foreignKey:InvestorID"`
// 	Comments      []Comment     `json:"comments,omitempty" gorm:"foreignKey:ObjectID"`
// 	Images        []ObjectImage `json:"images,omitempty" gorm:"foreignKey:ObjectID"`
// 	AverageRating *float64      `json:"averageRating" gorm:"-"`
// 	CommentCount  int           `json:"commentCount" gorm:"-"`
// }

type Comment struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ObjectID   uint      `json:"objectId" gorm:"not null"`
	AuthorName string    `json:"authorName" gorm:"not null"`
	Content    string    `json:"content" gorm:"not null"`
	Rating     *int      `json:"rating"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Statistic struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Label     string    `json:"label" gorm:"not null"`
	Value     string    `json:"value" gorm:"not null"`
	Icon      string    `json:"icon"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FAQ struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Question  string    `json:"question" gorm:"not null"`
	Answer    string    `json:"answer" gorm:"not null"`
	Category  string    `json:"category"`
	SortOrder int       `json:"sortOrder" gorm:"default:0"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type DashboardSummary struct {
	TotalObjects  int64   `json:"totalObjects"`
	ActiveObjects int64   `json:"activeObjects"`
	TotalOwners   int64   `json:"totalOwners"`
	TotalComments int64   `json:"totalComments"`
	TotalValue    float64 `json:"totalValue"`
	FeaturedCount int64   `json:"featuredCount"`
	AvgReturnRate float64 `json:"avgReturnRate"`
}


