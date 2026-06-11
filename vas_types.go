package squad

// BuyAirtimeParams holds parameters for purchasing airtime.
// PhoneNumber should be in international format without "+": "2348012345678".
// Minimum amount is 50 NGN. Charges come from the Squad wallet.
type BuyAirtimeParams struct {
	PhoneNumber    string `json:"phone_number"`
	Amount         int64  `json:"amount"`
	Network        string `json:"network_provider"` // "MTN", "AIRTEL", "GLO", "9MOBILE"
	TransactionRef string `json:"transaction_ref"`
}

// VASTransactionResponse is the common response for airtime, data, and cable purchases.
type VASTransactionResponse struct {
	TransactionRef string `json:"transaction_ref"`
	Amount         int64  `json:"amount"`
	Status         string `json:"status"`
	PhoneNumber    string `json:"phone_number,omitempty"`
	Network        string `json:"network,omitempty"`
	CreatedAt      string `json:"created_at"`
}

// DataPlan describes a single data bundle offering for a network provider.
type DataPlan struct {
	PlanCode string `json:"plan_code"`
	PlanName string `json:"plan_name"`
	Amount   int64  `json:"amount"`
	Validity string `json:"validity"`
	Network  string `json:"network"`
}

// DataPlansResponse holds available data plans for a network provider.
type DataPlansResponse struct {
	Plans []DataPlan `json:"plans"`
}

// BuyDataParams holds parameters for purchasing a data bundle.
type BuyDataParams struct {
	PhoneNumber    string `json:"phone_number"`
	PlanCode       string `json:"plan_code"`
	Network        string `json:"network_provider"`
	TransactionRef string `json:"transaction_ref"`
}

// CablePackage describes a single cable TV subscription package.
type CablePackage struct {
	PackageCode string `json:"package_code"`
	PackageName string `json:"package_name"`
	Amount      int64  `json:"amount"`
	CycleType   string `json:"cycle_type"` // "monthly", "quarterly"
}

// CablePackagesResponse holds available cable TV packages for a provider.
type CablePackagesResponse struct {
	Packages []CablePackage `json:"packages"`
}

// BuyCableParams holds parameters for subscribing to a cable TV package.
type BuyCableParams struct {
	SmartCardNumber string `json:"smart_card_number"`
	PackageCode     string `json:"package_code"`
	Provider        string `json:"cable_provider"` // "DSTV", "GOTV", "STARTIMES"
	TransactionRef  string `json:"transaction_ref"`
	Amount          int64  `json:"amount,omitempty"`
}

// ElectricityBiller describes a DISCO (electricity distribution company).
type ElectricityBiller struct {
	BillerCode string   `json:"biller_code"`
	BillerName string   `json:"biller_name"`
	MeterTypes []string `json:"meter_types"` // ["prepaid", "postpaid"]
}

// ElectricityBillersResponse holds the list of available electricity DISCOs.
type ElectricityBillersResponse struct {
	Billers []ElectricityBiller `json:"billers"`
}

// BuyElectricityParams holds parameters for purchasing electricity units.
type BuyElectricityParams struct {
	MeterNumber    string `json:"meter_number"`
	Amount         int64  `json:"amount"`
	BillerCode     string `json:"biller_code"`
	MeterType      string `json:"meter_type"` // "prepaid" or "postpaid"
	TransactionRef string `json:"transaction_ref"`
	PhoneNumber    string `json:"phone_number,omitempty"`
}

// ElectricityResponse is returned by BuyElectricity, including the token for the meter.
type ElectricityResponse struct {
	VASTransactionResponse
	MeterNumber      string `json:"meter_number"`
	Units            string `json:"units"`
	ElectricityToken string `json:"electricity_token"`
}

// SendSMSParams holds parameters for sending an SMS to one or more recipients.
type SendSMSParams struct {
	To             []string `json:"to"`
	From           string   `json:"from"`
	Body           string   `json:"body"`
	TransactionRef string   `json:"transaction_ref"`
}

// SMSResponse is returned by SendSMS.
type SMSResponse struct {
	TransactionRef string   `json:"transaction_ref"`
	Status         string   `json:"status"`
	Recipients     []string `json:"recipients"`
	MessageID      string   `json:"message_id"`
}
