package main

import (
	"encoding/xml"
	"regexp"
	"sync"
	"time"

	apiLib "github.com/hornbill/goApiLib"
)

// ----- Constants -----
const (
	version           = "3.5.0"
	repo              = "hornbill/goDBAssetImport"
	appServiceManager = "com.hornbill.servicemanager"
	appName           = "goDBAssetImport"
	maxGoRoutines     = 10
)

// ----- Variables -----
var (
	assets         = make(map[string]string)
	AssetClass     string
	AssetTypeID    int
	counters       counterTypeStruct
	logFilePart    = 0
	maxLogFileSize int64
	pageSize       int
	startTime      time.Time
	StrAssetType   string

	// DB variables
	BaseSQLQuery string
	connString   string
	importConf   importConfStruct
	key          keyDataStruct
	StrSQLAppend string

	// CLI argument variables
	configDebug        bool
	configDryRun       bool
	configFileName     string
	configForceUpdates bool
	configMaxRoutines  int
	configVersion      bool

	// Import config flags
	configCertero      bool
	configCSV          bool
	configDB           bool
	configGoogle       bool
	configLDAP         bool
	configNexthink     bool
	configWorkspaceOne bool

	// Global caches
	Sites                  []siteListStruct
	Groups                 []groupListStruct
	Customers              []customerListStruct
	HInstalledApplications = make(map[string]bool)

	// Worker stuff
	mutexAssets   = &sync.Mutex{}
	mutexBar      = &sync.Mutex{}
	mutexBuffer   = &sync.Mutex{}
	mutexCounters = &sync.Mutex{}
	worker        sync.WaitGroup

	// Shared Hornbill session for caching etc
	hornbillImport *apiLib.XmlmcInstStruct

	// Regex to check if a field contain Go templates
	regexTemplate, _ = regexp.Compile("{{.{1,}}}")
)

type counterTypeStruct struct {
	updated                            uint16
	created                            uint16
	updateSkipped                      uint16
	updateRelatedFailed                uint16
	updateRelatedSkipped               uint16
	createSkipped                      uint16
	updateFailed                       uint16
	createFailed                       uint16
	softwareCreated                    uint32
	softwareRemoved                    uint32
	softwareSkipped                    uint32
	softwareCreateFailed               uint32
	softwareRemoveFailed               uint32
	suppliersAssociatedSuccess         uint16
	suppliersAssociatedFailed          uint16
	suppliersAssociatedSkipped         uint16
	supplierContractsAssociatedSuccess uint16
	supplierContractsAssociatedFailed  uint16
	supplierContractsAssociatedSkipped uint16
}

// -- Cache Structs
type siteListStruct struct {
	SiteName string
	SiteID   int
}
type groupListStruct struct {
	GroupName string
	GroupType int
	GroupID   string
}
type customerListStruct struct {
	CustomerID     string
	UserID         string
	CustomerName   string
	CustomerHandle string
}

// -- Config Structs
type importConfStruct struct {
	APIKey                   string `json:"APIKey"`
	InstanceID               string `json:"InstanceId"`
	KeysafeKeyID             int    `json:"KeysafeKeyID"`
	AssetGenericFieldMapping map[string]interface{}
	AssetTypeFieldMapping    map[string]interface{}
	AssetTypes               []assetTypesStruct `json:"AssetTypes"`
	HornbillUserIDColumn     string             `json:"HornbillUserIDColumn"`
	LogSizeBytes             int64              `json:"LogSizeBytes"`
	SourceConfig             struct {
		CSV      csvConfStruct     `json:"CSV"`
		Database dbConfStruct      `json:"Database"`
		LDAP     ldapConfStruct    `json:"LDAP"`
		Google   googleConfStruct  `json:"Google"`
		Certero  certeroConfStruct `json:"Certero"`
		Source   string            `json:"Source"`
	} `json:"SourceConfig"`
}
type certeroConfStruct struct {
	Expand   string `json:"Expand"`
	PageSize int    `json:"PageSize"`
}
type csvConfStruct struct {
	CarriageReturnRemoval bool   `json:"CarriageReturnRemoval"`
	CommaCharacter        string `json:"CommaCharacter"`
	FieldsPerRecord       int    `json:"FieldsPerRecord"`
	LazyQuotes            bool   `json:"LazyQuotes"`
}
type dbConfStruct struct {
	Authentication string `json:"Authentication"`
	Encrypt        bool   `json:"Encrypt"`
	Query          string `json:"Query"`
}
type googleConfStruct struct {
	Customer    string `json:"Customer"`
	OrgUnitPath string `json:"OrgUnitPath"`
	Query       string `json:"Query"`
}
type ldapConfStruct struct {
	Query struct {
		Attributes   []string `json:"Attributes"`
		DerefAliases int      `json:"DerefAliases"`
		Scope        int      `json:"Scope"`
		SizeLimit    int      `json:"SizeLimit"`
		TimeLimit    int      `json:"TimeLimit"`
		TypesOnly    bool     `json:"TypesOnly"`
	} `json:"Query"`
	Server struct {
		ConnectionType     string `json:"ConnectionType"`
		Debug              bool   `json:"Debug"`
		InsecureSkipVerify bool   `json:"InsecureSkipVerify"`
	} `json:"Server"`
}
type assetTypesStruct struct {
	AssetIdentifier          assetIdentifierStruct   `json:"AssetIdentifier"`
	AssetType                string                  `json:"AssetType"`
	LDAPDSN                  string                  `json:"LDAPDSN"`
	NexthinkPlatform         string                  `json:"NexthinkPlatform"`
	CSVFile                  string                  `json:"CSVFile"`
	OperationType            string                  `json:"OperationType"`
	InPolicyField            string                  `json:"InPolicy"`
	PreserveOperationalState bool                    `json:"PreserveOperationalState"`
	PreserveShared           bool                    `json:"PreserveShared"`
	PreserveState            bool                    `json:"PreserveState"`
	PreserveSubState         bool                    `json:"PreserveSubState"`
	Query                    string                  `json:"Query"`
	SoftwareInventory        softwareInventoryStruct `json:"SoftwareInventory"`
	Class                    string                  `json:"Class"`
	TypeID                   int                     `json:"TypeID"`
	Filters                  filtersStruct           `json:"Filters"`
}
type filtersStruct struct {
	User                    string `json:"User"`
	Model_Identifier        string `json:"ModelIdentifier"`
	Device_Type             string `json:"DevicePlatformType"`
	Ownership               string `json:"Ownership"`
	Organization_Group_UUID string `json:"OrganizationGroupUUID"`
	Compliance_Status       string `json:"ComplianceStatus"`
	Seen_Since              string `json:"SeenSince"`
}
type assetIdentifierStruct struct {
	Entity               string `json:"Entity"`
	EntityColumn         string `json:"EntityColumn"`
	SourceColumn         string `json:"SourceColumn"`
	SourceContractColumn string `json:"SourceContractColumn"`
	SourceSupplierColumn string `json:"SourceSupplierColumn"`
}
type softwareInventoryStruct struct {
	AssetIDColumn string
	AppIDColumn   string
	Query         string
	Mapping       map[string]interface{}
	ParentObject  string
}
type keyDataStruct struct {
	AccessToken  string
	APIEndpoint  string `json:"api_endpoint"`
	APIKeyName   string `json:"apikeyname"`
	APIKey       string `json:"apikey"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Database     string `json:"database"`
	Domain       string `json:"domain"`
	Endpoint     string `json:"endpoint"`
	Host         string `json:"host"`
	Password     string `json:"password"`
	Port         uint16 `json:"port"`
	Region       string `json:"region"`
	Server       string `json:"server"`
	Username     string `json:"username"`
}

// -- XMLMC API Call Response Structs
type xmlmcResponse struct {
	MethodResult string         `xml:"status,attr"`
	Params       paramsStruct   `xml:"params"`
	State        stateStructXML `xml:"state"`
}
type stateStructXML struct {
	Code  string `xml:"code"`
	Error string `xml:"error"`
}
type stateStructJSON struct {
	Code      string `json:"code"`
	Service   string `json:"service"`
	Operation string `json:"operation"`
	Error     string `json:"error"`
}
type paramsStruct struct {
	SessionID               string `xml:"sessionId"`
	HPKID                   int    `xml:"primaryEntityData>record>h_pk_id"`
	Outcome                 string `xml:"outcome"`
	SupplierAssetID         int    `xml:"supplierAssetId"`
	SupplierContractAssetID int    `xml:"supplierContractAssetId"`
}
type xmlmcCountResponse struct {
	Params struct {
		RowData struct {
			Row []struct {
				Count string `json:"count"`
			} `json:"row"`
		} `json:"rowData"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
type xmlmcIBridgeResponse struct {
	MethodResult           string         `xml:"status,attr"`
	IBridgeResponsePayload string         `xml:"params>responsePayload"`
	IBridgeResponseError   string         `xml:"params>error"`
	State                  stateStructXML `xml:"state"`
}
type xmlmcKeyResponse struct {
	MethodResult string         `xml:"status,attr"`
	State        stateStructXML `xml:"state"`
	Params       struct {
		Data   string `json:"data"`
		Schema string `json:"schema"`
		Title  string `json:"title"`
		Type   string `json:"type"`
	} `json:"params"`
}
type xmlmcUpdateResponse struct {
	MethodResult string         `xml:"status,attr"`
	UpdatedCols  updatedCols    `xml:"params>primaryEntityData>record"`
	State        stateStructXML `xml:"state"`
}
type updatedCols struct {
	AssetPK string       `xml:"h_pk_asset_id"`
	ColList []updatedCol `xml:",any"`
}
type updatedCol struct {
	XMLName xml.Name `xml:""`
	Amount  string   `xml:",chardata"`
}
type xmlmcGroupResponse struct {
	Params struct {
		Group []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"group"`
		MaxPages int `json:"maxPages"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
type xmlmcAssetRecordsResponse struct {
	Params struct {
		RowData struct {
			Row []map[string]interface{} `json:"row"`
		} `json:"rowData"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
type xmlmcSoftwareRecordsResponse struct {
	Params struct {
		RowData struct {
			Row []softwareRecordDetailsStruct `xml:"row"`
		} `xml:"rowData"`
	} `xml:"params"`
	State stateStructJSON `xml:"state"`
}
type softwareRecordDetailsStruct struct {
	Count      uint64 `xml:"count"`
	HPKID      int    `xml:"h_pk_id"`
	HFKAssetID int    `xml:"h_fk_asset_id"`
	HAppName   string `xml:"h_app_name"`
	HAppID     string `xml:"h_app_id"`
}

// -- Sites Structs
type xmlmcSiteResponse struct {
	Params struct {
		Sites string `json:"sites"`
		Count int    `json:"count"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
type siteRowMultiple struct {
	Row []siteDetailsStruct `json:"row"`
}
type siteRowSingle struct {
	Row siteDetailsStruct `json:"row"`
}
type siteDetailsStruct struct {
	ID   string `json:"h_id"`
	Name string `json:"h_site_name"`
}

// -- User Structs
type xmlmcUserListResponse struct {
	Params struct {
		RowData struct {
			Row []userAccountStruct `json:"row"`
		} `json:"rowData"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
type userAccountStruct struct {
	HUserID     string `json:"h_user_id"`
	HLoginID    string `json:"h_login_id"`
	HEmployeeID string `json:"h_employee_id"`
	HName       string `json:"h_name"`
	HFirstName  string `json:"h_first_name"`
	HLastName   string `json:"h_last_name"`
	HEmail      string `json:"h_email"`
	HAttrib1    string `json:"h_attrib1"`
	HAttrib8    string `json:"h_attrib8"`
}

// -- Asset Type Structs
type xmlmcTypeListResponse struct {
	MethodResult string               `xml:"status,attr"`
	Params       paramsTypeListStruct `xml:"params"`
	State        stateStructXML       `xml:"state"`
}
type paramsTypeListStruct struct {
	Row assetTypeObjectStruct `xml:"rowData>row"`
}
type assetTypeObjectStruct struct {
	Type      string `xml:"h_name"`
	TypeClass string `xml:"h_class"`
	TypeID    int    `xml:"h_pk_type_id"`
}

// -- Application list structs
type xmlmcApplicationResponse struct {
	Status bool `json:"@status"`
	Params struct {
		Applications []struct {
			Name string `json:"name"`
		} `json:"application"`
	} `json:"params"`
	State stateStructJSON `json:"state"`
}
