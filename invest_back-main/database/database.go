package database

import (
	"investtrack-backend/models"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if os.Getenv("MIGRATE") == "true" {
		err = DB.AutoMigrate(
			&models.Owner{},
			&models.Investor{},
			&models.InvestmentObject{},
			&models.Comment{},
			&models.ObjectImage{},
			&models.Statistic{},
			&models.FAQ{},
		)
		if err != nil {
			log.Fatal("Failed to migrate database:", err)
		}
		log.Println("Database migration completed")
	}

	seed()
	migrateTopRatedObjects()
	migrateColumns()
	log.Println("Database initialized and seeded successfully")

	
	DB.AutoMigrate(
    &models.InvestmentObject{},
    &models.Owner{},
    &models.Investor{},
    &models.Comment{},
    &models.Statistic{},
    &models.FAQ{},
    &models.ObjectImage{}, // ← эта строчка должна быть
)
}

func seed() {
	var count int64
	DB.Model(&models.Owner{}).Count(&count)
	if count == 0 {
		seedCore()
	}
	seedImages()
}

// migrateColumns ensures new columns exist and backfills their values.
// This is a safety net for databases that were created before these columns
// were added to the model (e.g. the server was updated without re-running migrations).
func migrateColumns() {
	// Add top_rating column to investment_objects if it doesn't already exist
	DB.Exec(`ALTER TABLE investment_objects ADD COLUMN IF NOT EXISTS top_rating REAL DEFAULT 0`)

	// Add media_type column to object_images if it doesn't already exist
	DB.Exec(`ALTER TABLE object_images ADD COLUMN IF NOT EXISTS media_type TEXT DEFAULT 'image'`)

	// Backfill media_type for rows that are empty/null: detect video by file extension
	DB.Exec(`
		UPDATE object_images
		SET media_type = 'video'
		WHERE (media_type IS NULL OR media_type = '')
		  AND image_url ~* '\.(mp4|webm|mov|avi|mkv)$'
	`)

	// Remaining empty rows are images
	DB.Exec(`
		UPDATE object_images
		SET media_type = 'image'
		WHERE media_type IS NULL OR media_type = ''
	`)
}

func migrateTopRatedObjects() {
	var topRatedCount int64
	DB.Model(&models.InvestmentObject{}).Where("top_rated = ?", true).Count(&topRatedCount)
	if topRatedCount > 0 {
		return
	}

	DB.Model(&models.InvestmentObject{}).
		Where("featured = ?", true).
		Update("top_rated", true)
}

func seedCore() {

	owners := []models.Owner{
		{Name: "Alexander Richter", Company: "Richter Capital Group", Bio: "20+ years in commercial real estate across Europe and Asia. Harvard Business School graduate with $2B+ in transactions.", PhotoURL: "https://images.unsplash.com/photo-1560250097-0b93528c311a?w=400", Email: "a.richter@richter-capital.com", Phone: "+49 30 8844 2200"},
		{Name: "Sofia Marchetti", Company: "Marchetti Investments", Bio: "Pioneer in sustainable real estate development. Former Goldman Sachs VP. Specialist in ESG-certified properties.", PhotoURL: "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=400", Email: "sofia@marchetti-inv.com", Phone: "+39 02 7744 8800"},
		{Name: "James Chen", Company: "Pacific Bridge Capital", Bio: "Cross-border investment specialist focused on Asia-Pacific opportunities. $1.5B AUM.", PhotoURL: "https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?w=400", Email: "j.chen@pacificbridge.com", Phone: "+852 2244 6688"},
		{Name: "Natasha Volkov", Company: "Volkov Asset Management", Bio: "Eastern European market expert. Former Morgan Stanley Director. 15 years in emerging market real estate.", PhotoURL: "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=400", Email: "n.volkov@volkov-am.com", Phone: "+44 20 7946 0088"},
		{Name: "Omar Al-Rashid", Company: "Al-Rashid Holdings", Bio: "Gulf region investment authority with focus on luxury hospitality and mixed-use developments. 25 years experience.", PhotoURL: "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400", Email: "o.alrashid@ar-holdings.ae", Phone: "+971 4 388 9900"},
	}
	DB.Create(&owners)

	o1, o2, o3, o4, o5 := owners[0].ID, owners[1].ID, owners[2].ID, owners[3].ID, owners[4].ID
	val85, val120, val65, val200, val45, val175, val92, val38 := 85000000.0, 120000000.0, 65000000.0, 200000000.0, 45000000.0, 175000000.0, 92000000.0, 38000000.0

	objects := []models.InvestmentObject{
		{Title: "The Meridian Tower", Description: "A landmark 32-storey Grade-A commercial office tower in the heart of Frankfurt financial district. LEED Platinum certified with state-of-the-art infrastructure, attracting Fortune 500 tenants. Projected completion Q3 2026.", Type: "building", Status: "active", Location: "Frankfurt, Germany", ImageURL: "https://images.unsplash.com/photo-1486325212027-8081e485255e?w=1200", InvestmentValue: val85, ReturnRate: 8.4, Featured: true, OwnerID: &o1},
		{Title: "Riviera Business Park", Description: "Premium mixed-use business campus with 8 interconnected buildings across 12 hectares on the French Riviera. Fully leased to tech and finance tenants until 2031.", Type: "building", Status: "active", Location: "Nice, France", ImageURL: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=1200", InvestmentValue: val120, ReturnRate: 7.9, Featured: true, OwnerID: &o2},
		{Title: "GreenCore Logistics Hub", Description: "Next-generation sustainable logistics and distribution center with solar roof, EV fleet charging, and automated warehouse systems. Certified BREEAM Excellent.", Type: "construction", Status: "upcoming", Location: "Rotterdam, Netherlands", ImageURL: "https://images.unsplash.com/photo-1587293852726-70cdb56c2866?w=1200", InvestmentValue: val65, ReturnRate: 9.2, Featured: false, OwnerID: &o3},
		{Title: "Harbor Point Marina Complex", Description: "Iconic waterfront mixed-use development comprising luxury residences, retail promenade, superyacht marina, and boutique hotel. The crown jewel of Dubai's new coastal quarter.", Type: "building", Status: "active", Location: "Dubai, UAE", ImageURL: "https://images.unsplash.com/photo-1534430480872-3498386e7856?w=1200", InvestmentValue: val200, ReturnRate: 11.3, Featured: true, OwnerID: &o5},
		{Title: "TechHub Warsaw Campus", Description: "Purpose-built innovation campus for technology companies featuring co-working floors, R&D labs, auditorium and startup incubator. Warsaw's Silicon Valley.", Type: "building", Status: "active", Location: "Warsaw, Poland", ImageURL: "https://images.unsplash.com/photo-1497366754035-f200968a6e72?w=1200", InvestmentValue: val45, ReturnRate: 10.1, Featured: false, OwnerID: &o4},
		{Title: "Grand Palais Residences", Description: "Ultra-luxury residential tower on the banks of the Arno river, offering 48 exclusive apartments with private concierge, underground garage and rooftop infinity pool.", Type: "asset", Status: "completed", Location: "Florence, Italy", ImageURL: "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=1200", InvestmentValue: val175, ReturnRate: 6.8, Featured: true, OwnerID: &o2},
		{Title: "Silk Road Logistics Corridor", Description: "Strategic multi-modal logistics corridor with warehousing, customs processing and intermodal rail connections linking Central Asia to European markets.", Type: "construction", Status: "upcoming", Location: "Almaty, Kazakhstan", ImageURL: "https://images.unsplash.com/photo-1553413077-190dd305871c?w=1200", InvestmentValue: val92, ReturnRate: 12.5, Featured: false, OwnerID: &o3},
		{Title: "Alpine Wellness Resort", Description: "Five-star mountain wellness resort and spa with 180 rooms, thermal baths, ski-in/ski-out access and year-round event facilities. On-hold pending planning permits.", Type: "asset", Status: "on_hold", Location: "Zermatt, Switzerland", ImageURL: "https://images.unsplash.com/photo-1571896349842-33c89424de2d?w=1200", InvestmentValue: val38, ReturnRate: 7.5, Featured: false, OwnerID: &o1},
	}
	DB.Create(&objects)

	rating4, rating5, rating3 := 4, 5, 3
	comments := []models.Comment{
		{ObjectID: objects[0].ID, AuthorName: "Michael Brennan", Content: "Exceptional due diligence package from Richter Capital. The Frankfurt office market fundamentals are very strong right now, and this location is prime.", Rating: &rating5},
		{ObjectID: objects[0].ID, AuthorName: "Yuki Tanaka", Content: "Good asset with solid anchor tenants. Would have liked more granularity on lease break clauses, but overall a compelling risk-return profile.", Rating: &rating4},
		{ObjectID: objects[0].ID, AuthorName: "Pieter van den Berg", Content: "The LEED Platinum certification adds significant value for ESG-mandated institutional investors. Impressive sustainability credentials.", Rating: &rating4},
		{ObjectID: objects[1].ID, AuthorName: "Caroline Dupont", Content: "Riviera Business Park is the gold standard for campus-style office developments. Full occupancy through 2031 speaks for itself.", Rating: &rating5},
		{ObjectID: objects[1].ID, AuthorName: "Hartmut Schreiber", Content: "Strong covenant strength from the tech tenants. Infrastructure is future-proofed with fiber, power redundancy and EV charging.", Rating: &rating5},
		{ObjectID: objects[3].ID, AuthorName: "Fatima Al-Zahrawi", Content: "Harbor Point is a generational asset. Dubai Marina expansion is driving massive demand for premium waterfront developments.", Rating: &rating5},
		{ObjectID: objects[3].ID, AuthorName: "David Levi", Content: "The marina berth licensing gives this project a significant moat. Very few comparable waterfront sites remain available in Dubai.", Rating: &rating4},
		{ObjectID: objects[5].ID, AuthorName: "Lucia Romano", Content: "Grand Palais delivered exactly as promised. Top-tier finishes, incredible riverfront views. Sold all 48 units ahead of schedule.", Rating: &rating3},
	}
	DB.Create(&comments)

	stats := []models.Statistic{
		{Label: "Assets Under Management", Value: "$2.4B+", Icon: "DollarOutlined", SortOrder: 1},
		{Label: "Active Investments", Value: "156", Icon: "BuildOutlined", SortOrder: 2},
		{Label: "Countries", Value: "28", Icon: "GlobalOutlined", SortOrder: 3},
		{Label: "Average IRR", Value: "9.2%", Icon: "TrendingUpOutlined", SortOrder: 4},
		{Label: "Investor Partners", Value: "1,200+", Icon: "TeamOutlined", SortOrder: 5},
		{Label: "Years Experience", Value: "18", Icon: "TrophyOutlined", SortOrder: 6},
	}
	DB.Create(&stats)

	faqs := []models.FAQ{
		{Question: "What is the minimum investment amount?", Answer: "Our minimum investment threshold varies by asset class. Commercial properties typically start at $250,000, while our flagship fund products have a minimum of $1,000,000. Contact our investor relations team for personalized guidance.", Category: "Investment", SortOrder: 1},
		{Question: "How are returns distributed to investors?", Answer: "Returns are distributed quarterly for income-producing assets and upon liquidity events for development projects. All distributions are processed electronically to your nominated bank account within 5 business days of quarter-end.", Category: "Returns", SortOrder: 2},
		{Question: "What due diligence process do you follow?", Answer: "We conduct a rigorous 90-day due diligence process covering legal title review, environmental assessments, structural surveys, market analysis, financial modeling, and independent valuation. All assets are stress-tested against multiple economic scenarios.", Category: "Process", SortOrder: 3},
		{Question: "What types of assets do you invest in?", Answer: "We focus on three core asset classes: Grade-A commercial office and retail properties, logistics and industrial facilities, and premium residential developments. All assets must meet our ESG criteria and minimum return thresholds.", Category: "Assets", SortOrder: 4},
		{Question: "How is my investment protected?", Answer: "Investments are held in bankruptcy-remote special purpose vehicles (SPVs). We maintain comprehensive property and liability insurance, employ conservative LTV ratios below 60%, and hold cash reserves equivalent to 12 months of operating expenses.", Category: "Security", SortOrder: 5},
		{Question: "Can international investors participate?", Answer: "Yes, we welcome qualified international investors from over 40 countries. We provide fully digital onboarding with KYC/AML compliance, multi-currency subscription, and dedicated cross-border tax structuring advice.", Category: "Eligibility", SortOrder: 6},
		{Question: "What reporting do investors receive?", Answer: "Investors receive quarterly NAV statements, monthly portfolio performance updates, annual audited accounts, and real-time access to our investor portal. All documentation is available in English, German, French, and Arabic.", Category: "Reporting", SortOrder: 7},
		{Question: "What is the typical investment horizon?", Answer: "Investment horizons vary by asset type. Core income assets typically have a 5–7 year hold period, value-add projects 3–5 years, and development assets 2–4 years from initial close. We target 12-month liquidity windows for all strategies.", Category: "Investment", SortOrder: 8},
	}
	DB.Create(&faqs)
}

func seedImages() {
	var imgCount int64
	DB.Model(&models.ObjectImage{}).Count(&imgCount)
	if imgCount > 0 {
		return
	}

	var objects []models.InvestmentObject
	DB.Find(&objects)
	if len(objects) == 0 {
		return
	}

	idMap := map[string]uint{}
	for _, o := range objects {
		idMap[o.Title] = o.ID
	}

	images := []models.ObjectImage{}

	// The Meridian Tower — Frankfurt office tower
	if id, ok := idMap["The Meridian Tower"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1486325212027-8081e485255e?w=1200", Caption: "Main facade — Frankfurt financial district", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1554435493-93422e8220c8?w=1200", Caption: "Grade-A lobby interior", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1497366811353-6870744d04b2?w=1200", Caption: "Open-plan office floors", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1497366754035-f200968a6e72?w=1200", Caption: "Executive conference suite", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1580587771525-78b9dba3b914?w=1200", Caption: "Rooftop terrace with city views", SortOrder: 5},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?w=1200", Caption: "Underground parking & EV charging", SortOrder: 6},
		)
	}

	// Riviera Business Park — Nice, France
	if id, ok := idMap["Riviera Business Park"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1497366216548-37526070297c?w=1200", Caption: "Campus aerial view — French Riviera", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1504384308090-c894fdcc538d?w=1200", Caption: "Co-working innovation hub", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1556761175-5973dc0f32e7?w=1200", Caption: "Outdoor meeting terraces", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1600880292203-757bb62b4baf?w=1200", Caption: "Landscaped common areas", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1542744173-8e7e53415bb0?w=1200", Caption: "Executive boardroom", SortOrder: 5},
		)
	}

	// GreenCore Logistics Hub — Rotterdam
	if id, ok := idMap["GreenCore Logistics Hub"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1587293852726-70cdb56c2866?w=1200", Caption: "Distribution centre overview", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1553413077-190dd305871c?w=1200", Caption: "Automated warehouse interior", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?w=1200", Caption: "Solar roof installation", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=1200", Caption: "EV fleet charging station", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1519999482648-25049ddd37b1?w=1200", Caption: "Loading dock — Rotterdam port access", SortOrder: 5},
		)
	}

	// Harbor Point Marina Complex — Dubai
	if id, ok := idMap["Harbor Point Marina Complex"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1534430480872-3498386e7856?w=1200", Caption: "Waterfront development — Dubai Marina", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1512453979798-5ea266f8880c?w=1200", Caption: "Superyacht marina berths", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1545324418-cc1a3fa10c00?w=1200", Caption: "Luxury residential tower", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1582719478250-c89cae4dc85b?w=1200", Caption: "Boutique hotel infinity pool", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1578683010236-d716f9a3f461?w=1200", Caption: "Retail promenade at sunset", SortOrder: 5},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1590490360182-c33d57733427?w=1200", Caption: "Penthouse living area", SortOrder: 6},
		)
	}

	// TechHub Warsaw Campus
	if id, ok := idMap["TechHub Warsaw Campus"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1497366754035-f200968a6e72?w=1200", Caption: "Campus entrance — Warsaw tech district", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1504384308090-c894fdcc538d?w=1200", Caption: "Open co-working floors", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1519389950473-47ba0277781c?w=1200", Caption: "R&D laboratory wing", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=1200", Caption: "Startup incubator auditorium", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1497366811353-6870744d04b2?w=1200", Caption: "Private meeting pods", SortOrder: 5},
		)
	}

	// Grand Palais Residences — Florence
	if id, ok := idMap["Grand Palais Residences"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1512917774080-9991f1c4c750?w=1200", Caption: "Riverfront facade — Florence", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?w=1200", Caption: "Grand lobby with marble finishes", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?w=1200", Caption: "Signature penthouse interior", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1600607687644-c7171b42498b?w=1200", Caption: "Rooftop infinity pool — Arno views", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1613977257363-707ba9348227?w=1200", Caption: "Private concierge lounge", SortOrder: 5},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1574362848149-11496d93a7c7?w=1200", Caption: "Underground private garage", SortOrder: 6},
		)
	}

	// Silk Road Logistics Corridor — Almaty
	if id, ok := idMap["Silk Road Logistics Corridor"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1553413077-190dd305871c?w=1200", Caption: "Logistics corridor — Almaty hub", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1587293852726-70cdb56c2866?w=1200", Caption: "Intermodal rail terminal", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1519999482648-25049ddd37b1?w=1200", Caption: "Customs processing facility", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1601584115197-04ecc0da31d7?w=1200", Caption: "Cold storage warehouse block", SortOrder: 4},
		)
	}

	// Alpine Wellness Resort — Zermatt
	if id, ok := idMap["Alpine Wellness Resort"]; ok {
		images = append(images,
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1571896349842-33c89424de2d?w=1200", Caption: "Mountain resort exterior — Zermatt", SortOrder: 1},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1540555700478-4be289fbecef?w=1200", Caption: "Thermal spa & wellness centre", SortOrder: 2},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1566073771259-6a8506099945?w=1200", Caption: "Ski-in/ski-out lodge area", SortOrder: 3},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1520250497591-112f2f40a3f4?w=1200", Caption: "Panoramic restaurant & bar", SortOrder: 4},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1582719478250-c89cae4dc85b?w=1200", Caption: "Suite interior with Matterhorn view", SortOrder: 5},
			models.ObjectImage{ObjectID: id, ImageURL: "https://images.unsplash.com/photo-1584132967334-10e028bd69f7?w=1200", Caption: "Event hall — 400 capacity", SortOrder: 6},
		)
	}

	if len(images) > 0 {
		DB.Create(&images)
	}
}
