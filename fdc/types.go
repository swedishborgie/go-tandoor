package fdc

import (
	"encoding/json"
	"strconv"
)

// FlexibleInt can unmarshal a JSON value that is either an int or a string
// representation of an int. The FDC API returns ndbNumber as int in some
// endpoints and string in others; this type handles both transparently.
type FlexibleInt struct {
	Value    int
	HasValue bool
}

// UnmarshalJSON handles both integer and string JSON values.
func (f *FlexibleInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.Value = 0
		f.HasValue = false
		return nil
	}
	// Try int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		f.Value = i
		f.HasValue = true
		return nil
	}
	// Fall back to string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		f.Value = v
		f.HasValue = true
		return nil
	}
	return &json.UnmarshalTypeError{
		Value: string(data),
		Type:  nil,
	}
}

// MarshalJSON produces a plain JSON number or null.
func (f FlexibleInt) MarshalJSON() ([]byte, error) {
	if !f.HasValue {
		return []byte("null"), nil
	}
	return []byte(strconv.Itoa(f.Value)), nil
}

// String returns the decimal string representation or empty string if null.
func (f FlexibleInt) String() string {
	if !f.HasValue {
		return ""
	}
	return strconv.Itoa(f.Value)
}

// FlexibleString can unmarshal a JSON value that is either a string or a
// number. The FDC API returns modifier as an integer in some foods and a
// string in others; this type handles both transparently.
type FlexibleString struct {
	Value    string
	HasValue bool
}

// UnmarshalJSON handles both string and numeric JSON values.
func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		f.Value = ""
		f.HasValue = false
		return nil
	}
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		f.Value = s
		f.HasValue = true
		return nil
	}
	// Fall back to number
	var n float64
	if err := json.Unmarshal(data, &n); err != nil {
		return &json.UnmarshalTypeError{
			Value: string(data),
			Type:  nil,
		}
	}
	f.Value = strconv.FormatFloat(n, 'f', -1, 64)
	f.HasValue = true
	return nil
}

// MarshalJSON produces a plain JSON string or null.
func (f FlexibleString) MarshalJSON() ([]byte, error) {
	if !f.HasValue {
		return []byte("null"), nil
	}
	return json.Marshal(f.Value)
}

// String returns the value or empty string if null.
func (f FlexibleString) String() string {
	return f.Value
}

// Format controls how much nutrient detail is returned.
type Format string

const (
	// FormatAbridged returns abridged nutrient data.
	FormatAbridged Format = "abridged"
	// FormatFull returns full nutrient data.
	FormatFull Format = "full"
)

// DataType filters food results by data type.
type DataType string

const (
	// DataTypeBranded filters for branded foods.
	DataTypeBranded DataType = "Branded"
	// DataTypeFoundation filters for foundation foods.
	DataTypeFoundation DataType = "Foundation"
	// DataTypeSurvey filters for survey foods.
	DataTypeSurvey DataType = "Survey (FNDDS)"
	// DataTypeSRLegacy filters for SR Legacy foods.
	DataTypeSRLegacy DataType = "SR Legacy"
)

// SortField specifies the sort field for listing and search.
type SortField string

const (
	// SortFieldDataType sorts by data type.
	SortFieldDataType SortField = "dataType.keyword"
	// SortFieldDescription sorts by description.
	SortFieldDescription SortField = "lowercaseDescription.keyword"
	// SortFieldFdcID sorts by FDC ID.
	SortFieldFdcID SortField = "fdcId"
	// SortFieldPublishedDate sorts by published date.
	SortFieldPublishedDate SortField = "publishedDate"
)

// SortOrder specifies ascending or descending sort order.
type SortOrder string

const (
	// SortOrderAsc sorts ascending.
	SortOrderAsc SortOrder = "asc"
	// SortOrderDesc sorts descending.
	SortOrderDesc SortOrder = "desc"
)

// Food is a unified food item that can represent any food type returned by
// the FDC API (Branded, Foundation, SR Legacy, Survey). Optional fields
// are pointers so they can be distinguished from zero values.
type Food struct {
	FDCID                     int                  `json:"fdcId,omitempty"`
	DataType                  string               `json:"dataType,omitempty"`
	Description               string               `json:"description,omitempty"`
	FoodClass                 string               `json:"foodClass,omitempty"`
	PublicationDate           string               `json:"publicationDate,omitempty"`
	AvailableDate             string               `json:"availableDate,omitempty"`
	ModifiedDate              string               `json:"modifiedDate,omitempty"`
	StartDate                 string               `json:"startDate,omitempty"`
	EndDate                   string               `json:"endDate,omitempty"`
	BrandOwner                string               `json:"brandOwner,omitempty"`
	GTINUPC                   string               `json:"gtinUpc,omitempty"`
	NDBNumber                 FlexibleInt          `json:"ndbNumber,omitempty"`
	FoodCode                  FlexibleInt          `json:"foodCode,omitempty"`
	DataSource                string               `json:"dataSource,omitempty"`
	FoodCategory              *FoodCategory        `json:"foodCategory,omitempty"`
	BrandedFoodCategory       string               `json:"brandedFoodCategory,omitempty"`
	ServingSize               *float64             `json:"servingSize,omitempty"`
	ServingSizeUnit           string               `json:"servingSizeUnit,omitempty"`
	HouseholdServingFullText  string               `json:"householdServingFullText,omitempty"`
	Ingredients               string               `json:"ingredients,omitempty"`
	ScientificName            string               `json:"scientificName,omitempty"`
	IsHistoricalReference     *bool                `json:"isHistoricalReference,omitempty"`
	FootNote                  string               `json:"footNote,omitempty"`
	LabelNutrients            *LabelNutrients      `json:"labelNutrients,omitempty"`
	FoodNutrients             []FoodNutrient       `json:"foodNutrients,omitempty"`
	FoodAttributes            []FoodAttribute      `json:"foodAttributes,omitempty"`
	FoodPortions              []FoodPortion        `json:"foodPortions,omitempty"`
	FoodComponents            []FoodComponent      `json:"foodComponents,omitempty"`
	InputFoods                []InputFood          `json:"inputFoods,omitempty"`
	NutrientConversionFactors []NutrientConversion `json:"nutrientConversionFactors,omitempty"`
	FoodUpdateLog             []FoodUpdateLog      `json:"foodUpdateLog,omitempty"`
	WweiaFoodCategory         *WweiaFoodCategory   `json:"wweiaFoodCategory,omitempty"`
}

// AbridgedFood is the reduced food item returned by list and search endpoints.
type AbridgedFood struct {
	FDCID           int                    `json:"fdcId"`
	DataType        string                 `json:"dataType"`
	Description     string                 `json:"description"`
	BrandOwner      string                 `json:"brandOwner,omitempty"`
	GTINUPC         string                 `json:"gtinUpc,omitempty"`
	NDBNumber       FlexibleInt            `json:"ndbNumber,omitempty"`
	FoodCode        FlexibleInt            `json:"foodCode,omitempty"`
	PublicationDate string                 `json:"publicationDate,omitempty"`
	FoodNutrients   []AbridgedFoodNutrient `json:"foodNutrients,omitempty"`
}

// SearchFood is a food item as returned within a search result.
type SearchFood struct {
	FDCID                  int                    `json:"fdcId"`
	DataType               string                 `json:"dataType,omitempty"`
	Description            string                 `json:"description"`
	FoodCode               FlexibleInt            `json:"foodCode,omitempty"`
	BrandOwner             string                 `json:"brandOwner,omitempty"`
	GTINUPC                string                 `json:"gtinUpc,omitempty"`
	Ingredients            string                 `json:"ingredients,omitempty"`
	NDBNumber              FlexibleInt            `json:"ndbNumber,omitempty"`
	PublicationDate        string                 `json:"publicationDate,omitempty"`
	ScientificName         string                 `json:"scientificName,omitempty"`
	AdditionalDescriptions string                 `json:"additionalDescriptions,omitempty"`
	AllHighlightFields     string                 `json:"allHighlightFields,omitempty"`
	Score                  *float64               `json:"score,omitempty"`
	FoodNutrients          []AbridgedFoodNutrient `json:"foodNutrients,omitempty"`
}

// SearchResponse wraps search results with pagination metadata.
type SearchResponse struct {
	FoodSearchCriteria any          `json:"foodSearchCriteria,omitempty"`
	TotalHits          int          `json:"totalHits"`
	CurrentPage        int          `json:"currentPage"`
	TotalPages         int          `json:"totalPages"`
	Foods              []SearchFood `json:"foods"`
}

// --- Nutrient types ---

// FoodNutrient contains detailed nutrient data for a food.
type FoodNutrient struct {
	ID                      uint               `json:"id,omitempty"`
	Amount                  *float64           `json:"amount,omitempty"`
	DataPoints              *int               `json:"dataPoints,omitempty"`
	Min                     *float64           `json:"min,omitempty"`
	Max                     *float64           `json:"max,omitempty"`
	Median                  *float64           `json:"median,omitempty"`
	Type                    string             `json:"type,omitempty"`
	Nutrient                *Nutrient          `json:"nutrient,omitempty"`
	FoodNutrientDerivation  *NutrientDeriv     `json:"foodNutrientDerivation,omitempty"`
	NutrientAnalysisDetails []NutrientAnalysis `json:"nutrientAnalysisDetails,omitempty"`
}

// AbridgedFoodNutrient is the reduced nutrient record in abridged responses.
type AbridgedFoodNutrient struct {
	Number                uint     `json:"number,omitempty"`
	Name                  string   `json:"name,omitempty"`
	Amount                *float64 `json:"amount,omitempty"`
	UnitName              string   `json:"unitName,omitempty"`
	DerivationCode        string   `json:"derivationCode,omitempty"`
	DerivationDescription string   `json:"derivationDescription,omitempty"`
}

// Nutrient describes a nutrient definition.
type Nutrient struct {
	ID       uint   `json:"id,omitempty"`
	Number   string `json:"number,omitempty"`
	Name     string `json:"name,omitempty"`
	Rank     *uint  `json:"rank,omitempty"`
	UnitName string `json:"unitName,omitempty"`
}

// NutrientDeriv describes how a nutrient value was derived.
type NutrientDeriv struct {
	ID                 int                 `json:"id,omitempty"`
	Code               string              `json:"code,omitempty"`
	Description        string              `json:"description,omitempty"`
	FoodNutrientSource *FoodNutrientSource `json:"foodNutrientSource,omitempty"`
}

// FoodNutrientSource describes the source of nutrient data.
type FoodNutrientSource struct {
	ID          int    `json:"id,omitempty"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// NutrientAnalysis contains lab-level nutrient analysis details.
type NutrientAnalysis struct {
	SubSampleID                  int                   `json:"subSampleId,omitempty"`
	Amount                       *float64              `json:"amount,omitempty"`
	NutrientID                   int                   `json:"nutrientId,omitempty"`
	LabMethodDescription         string                `json:"labMethodDescription,omitempty"`
	LabMethodOriginalDescription string                `json:"labMethodOriginalDescription,omitempty"`
	LabMethodLink                string                `json:"labMethodLink,omitempty"`
	LabMethodTechnique           string                `json:"labMethodTechnique,omitempty"`
	NutrientAcquisitionDetails   []NutrientAcquisition `json:"nutrientAcquisitionDetails,omitempty"`
}

// NutrientAcquisition describes how a nutrient sample was acquired.
type NutrientAcquisition struct {
	SampleUnitID int    `json:"sampleUnitId,omitempty"`
	PurchaseDate string `json:"purchaseDate,omitempty"`
	StoreCity    string `json:"storeCity,omitempty"`
	StoreState   string `json:"storeState,omitempty"`
}

// --- Food component types ---

// FoodPortion describes a serving/portion of a food.
type FoodPortion struct {
	ID                 uint           `json:"id,omitempty"`
	Amount             *float64       `json:"amount,omitempty"`
	DataPoints         *int           `json:"dataPoints,omitempty"`
	GramWeight         *float64       `json:"gramWeight,omitempty"`
	MinYearAcquired    int            `json:"minYearAcquired,omitempty"`
	Modifier           FlexibleString `json:"modifier,omitempty"`
	PortionDescription string         `json:"portionDescription,omitempty"`
	SequenceNumber     int            `json:"sequenceNumber,omitempty"`
	MeasureUnit        *MeasureUnit   `json:"measureUnit,omitempty"`
}

// MeasureUnit describes a unit of measure.
type MeasureUnit struct {
	ID           uint   `json:"id,omitempty"`
	Abbreviation string `json:"abbreviation,omitempty"`
	Name         string `json:"name,omitempty"`
}

// FoodComponent describes a component part of a food.
type FoodComponent struct {
	ID              uint     `json:"id,omitempty"`
	Name            string   `json:"name,omitempty"`
	DataPoints      int      `json:"dataPoints,omitempty"`
	GramWeight      *float64 `json:"gramWeight,omitempty"`
	IsRefuse        *bool    `json:"isRefuse,omitempty"`
	MinYearAcquired int      `json:"minYearAcquired,omitempty"`
	PercentWeight   *float64 `json:"percentWeight,omitempty"`
}

// FoodCategory describes a food category.
type FoodCategory struct {
	ID          int    `json:"id,omitempty"`
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// FoodAttribute describes a metadata attribute on a food.
// Value is any because the FDC API returns it as string, int, or null across different foods.
type FoodAttribute struct {
	ID                int            `json:"id,omitempty"`
	SequenceNumber    int            `json:"sequenceNumber,omitempty"`
	Value             any            `json:"value,omitempty"`
	FoodAttributeType *AttributeType `json:"FoodAttributeType,omitempty"`
}

// AttributeType describes the type of a food attribute.
type AttributeType struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// InputFood describes an input food (ingredient) for composite foods.
type InputFood struct {
	ID                    uint             `json:"id,omitempty"`
	FoodDescription       string           `json:"foodDescription,omitempty"`
	Amount                *float64         `json:"amount,omitempty"`
	IngredientCode        int              `json:"ingredientCode,omitempty"`
	IngredientDescription string           `json:"ingredientDescription,omitempty"`
	IngredientWeight      *float64         `json:"ingredientWeight,omitempty"`
	PortionCode           string           `json:"portionCode,omitempty"`
	PortionDescription    string           `json:"portionDescription,omitempty"`
	SequenceNumber        int              `json:"sequenceNumber,omitempty"`
	SurveyFlag            int              `json:"surveyFlag,omitempty"`
	Unit                  string           `json:"unit,omitempty"`
	RetentionFactor       *RetentionFactor `json:"retentionFactor,omitempty"`
	InputFood             any              `json:"inputFood,omitempty"`
}

// RetentionFactor describes nutrient retention during cooking.
type RetentionFactor struct {
	ID          int    `json:"id,omitempty"`
	Code        int    `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// NutrientConversion describes a nutrient conversion factor.
type NutrientConversion struct {
	Type  string   `json:"type,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

// WweiaFoodCategory describes a WWEIA food category (Survey foods).
type WweiaFoodCategory struct {
	WweiaFoodCategoryCode        int    `json:"wweiaFoodCategoryCode,omitempty"`
	WweiaFoodCategoryDescription string `json:"wweiaFoodCategoryDescription,omitempty"`
}

// --- Label nutrients ---

// LabelNutrients contains nutrition label values.
type LabelNutrients struct {
	Fat           *LabelNutrientValue `json:"fat,omitempty"`
	SaturatedFat  *LabelNutrientValue `json:"saturatedFat,omitempty"`
	TransFat      *LabelNutrientValue `json:"transFat,omitempty"`
	Cholesterol   *LabelNutrientValue `json:"cholesterol,omitempty"`
	Sodium        *LabelNutrientValue `json:"sodium,omitempty"`
	Carbohydrates *LabelNutrientValue `json:"carbohydrates,omitempty"`
	Fiber         *LabelNutrientValue `json:"fiber,omitempty"`
	Sugars        *LabelNutrientValue `json:"sugars,omitempty"`
	Protein       *LabelNutrientValue `json:"protein,omitempty"`
	Calcium       *LabelNutrientValue `json:"calcium,omitempty"`
	Iron          *LabelNutrientValue `json:"iron,omitempty"`
	Potassium     *LabelNutrientValue `json:"postassium,omitempty"`
	Calories      *LabelNutrientValue `json:"calories,omitempty"`
}

// LabelNutrientValue is a simple value wrapper for label nutrients.
type LabelNutrientValue struct {
	Value *float64 `json:"value,omitempty"`
}

// FoodUpdateLog describes a historical update to a food record.
type FoodUpdateLog struct {
	FDCID                    int             `json:"fdcId,omitempty"`
	AvailableDate            string          `json:"availableDate,omitempty"`
	BrandOwner               string          `json:"brandOwner,omitempty"`
	DataSource               string          `json:"dataSource,omitempty"`
	DataType                 string          `json:"dataType,omitempty"`
	Description              string          `json:"description,omitempty"`
	FoodClass                string          `json:"foodClass,omitempty"`
	GTINUPC                  string          `json:"gtinUpc,omitempty"`
	HouseholdServingFullText string          `json:"householdServingFullText,omitempty"`
	Ingredients              string          `json:"ingredients,omitempty"`
	ModifiedDate             string          `json:"modifiedDate,omitempty"`
	PublicationDate          string          `json:"publicationDate,omitempty"`
	ServingSize              *float64        `json:"servingSize,omitempty"`
	ServingSizeUnit          string          `json:"servingSizeUnit,omitempty"`
	BrandedFoodCategory      string          `json:"brandedFoodCategory,omitempty"`
	Changes                  string          `json:"changes,omitempty"`
	FoodAttributes           []FoodAttribute `json:"foodAttributes,omitempty"`
}

// --- Request criteria types ---

// FoodsCriteria is the request body for POST /v1/foods.
type FoodsCriteria struct {
	FDCIDs    []int   `json:"fdcIds"`
	Format    *Format `json:"format,omitempty"`
	Nutrients []int   `json:"nutrients,omitempty"`
}

// FoodListCriteria is the request body for POST /v1/foods/list.
type FoodListCriteria struct {
	DataType   []DataType `json:"dataType,omitempty"`
	PageSize   *int       `json:"pageSize,omitempty"`
	PageNumber *int       `json:"pageNumber,omitempty"`
	SortBy     *SortField `json:"sortBy,omitempty"`
	SortOrder  *SortOrder `json:"sortOrder,omitempty"`
}

// FoodSearchCriteria is the request body for POST /v1/foods/search.
type FoodSearchCriteria struct {
	Query      string     `json:"query"`
	DataType   []DataType `json:"dataType,omitempty"`
	PageSize   *int       `json:"pageSize,omitempty"`
	PageNumber *int       `json:"pageNumber,omitempty"`
	SortBy     *SortField `json:"sortBy,omitempty"`
	SortOrder  *SortOrder `json:"sortOrder,omitempty"`
	BrandOwner *string    `json:"brandOwner,omitempty"`
}
