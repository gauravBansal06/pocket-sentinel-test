package common

type ProcessInfo struct {
	PID               int
	Name              string
	IsApplication     bool
	BundleIdentifier  string
	ForegroundRunning bool
}

type TestInfo struct {
	OS             string `json:"os"`
	UDID           string `json:"udid"`
	TestType       string `json:"testType"`
	AppPackage     string `json:"appPackage"`
	AppActivity    string `json:"appActivity"`
	TestID         string `json:"testId"`
	SessionID      string `json:"sessionId"`
	Version        string `json:"version"`
	AppPath        string `json:"appPath"`
	HubURL         string `json:"hubUrl"`
	AutomationName string `json:"automationName"`
	VideoLogs      string `json:"videoLogs"`
	AppiumLogs     string `json:"appiumLogs"`
	CommandLogs    string `json:"commandLogs"`
	Value          struct {
		SessionID string `json:"sessionId"`
	} `json:"value"`
}

type WebDriver struct {
	Capabilities        Capabilities        `json:"capabilities"`
	DesiredCapabilities DesiredCapabilities `json:"desiredCapabilities"`
}

type Capabilities struct {
	AlwaysMatch AlwaysMatch `json:"alwaysMatch"`
	FirstMatch  []struct{}  `json:"firstMatch"`
}

type AlwaysMatch struct {
	AppiumUDID                    string `json:"appium:udid"`
	PlatformName                  string `json:"platformName"`
	AppiumAutomationName          string `json:"appium:automationName"`
	AppiumNoReset                 bool   `json:"appium:noReset"`
	AppiumEnsureWebviewsHavePages bool   `json:"appium:ensureWebviewsHavePages"`
	AppiumNativeWebScreenshot     bool   `json:"appium:nativeWebScreenshot"`
	AppiumNewCommandTimeout       int    `json:"appium:newCommandTimeout"`
	AppiumConnectHardwareKeyboard bool   `json:"appium:connectHardwareKeyboard"`
}

type DesiredCapabilities struct {
	AppiumUDID                    string `json:"appium:udid"`
	PlatformName                  string `json:"platformName"`
	AutomationName                string `json:"automationName"`
	AppiumNoReset                 bool   `json:"appium:noReset"`
	AppiumAppPackage              string `json:"appium:appPackage"`
	AppiumAppActivity             string `json:"appium:appActivity"`
	AppiumEnsureWebviewsHavePages bool   `json:"appium:ensureWebviewsHavePages"`
	AppiumNativeWebScreenshot     bool   `json:"appium:nativeWebScreenshot"`
	AppiumNewCommandTimeout       int    `json:"appium:newCommandTimeout"`
	AppiumConnectHardwareKeyboard bool   `json:"appium:connectHardwareKeyboard"`
}

type Organization struct {
	OrgID          int         `json:"id"`
	Name           string      `json:"name"`
	PlanAttributes interface{} `json:"plan_attributes"`
}

type UserDetails struct {
	UserID       int          `json:"id"`
	Name         string       `json:"name"`
	Email        string       `json:"email"`
	Username     string       `json:"username"`
	Status       string       `json:"status"`
	Role         string       `json:"organization_role"`
	OrgID        int          `json:"org_id"`
	ApiToken     string       `json:"apiToken"`
	Organization Organization `json:"organization"`
}

type BearerAuthUserDetails struct {
	UserID   int    `json:"userID"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Status   string `json:"status"`
	Role     string `json:"role"`
	Token    string `json:"token"`
	OrgID    int    `json:"orgID"`
}
