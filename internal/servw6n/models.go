package models

type Config struct {
	RPID             string `json:"RPID,omitempty"`
	RPDisplayName    string `json:"RPDisplayName,omitempty"`
	RPOrigin         string `json:"RPOrigin,omitempty"`
	UserVerification string `json:"UserVerification,omitempty"`
}

func (m *Config) GetRPID() string {
	if m != nil {
		return m.RPID
	}
	return ""
}

func (m *Config) GetRPDisplayName() string {
	if m != nil {
		return m.RPDisplayName
	}
	return ""
}

func (m *Config) GetRPOrigin() string {
	if m != nil {
		return m.RPOrigin
	}
	return ""
}

func (m *Config) GetUserVerification() string {
	if m != nil {
		return m.UserVerification
	}
	return ""
}

type User struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func (m *User) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *User) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type EnrollInitReq struct {
	User     *User   `json:"user,omitempty"`
	UserBlob []byte  `json:"userBlob,omitempty"`
	Cfg      *Config `json:"cfg,omitempty"`
}

func (m *EnrollInitReq) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

func (m *EnrollInitReq) GetUserBlob() []byte {
	if m != nil {
		return m.UserBlob
	}
	return nil
}

func (m *EnrollInitReq) GetCfg() *Config {
	if m != nil {
		return m.Cfg
	}
	return nil
}

type AuthInitReq struct {
	UserBlob []byte  `json:"userBlob,omitempty"`
	Cfg      *Config `json:"cfg,omitempty"`
}

func (m *AuthInitReq) GetUserBlob() []byte {
	if m != nil {
		return m.UserBlob
	}
	return nil
}

func (m *AuthInitReq) GetCfg() *Config {
	if m != nil {
		return m.Cfg
	}
	return nil
}

type InitRes struct {
	Session []byte `json:"session,omitempty"`
	Json    []byte `json:"json,omitempty"`
}

func (m *InitRes) GetSession() []byte {
	if m != nil {
		return m.Session
	}
	return nil
}

func (m *InitRes) GetJson() []byte {
	if m != nil {
		return m.Json
	}
	return nil
}

type FinalReq struct {
	Session   []byte  `json:"session,omitempty"`
	Signature []byte  `json:"signature,omitempty"`
	Cfg       *Config `json:"cfg,omitempty"`
}

func (m *FinalReq) GetSession() []byte {
	if m != nil {
		return m.Session
	}
	return nil
}

func (m *FinalReq) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *FinalReq) GetCfg() *Config {
	if m != nil {
		return m.Cfg
	}
	return nil
}

type FinalRes struct {
	Valid    bool   `json:"valid,omitempty"`
	UserBlob []byte `json:"userBlob,omitempty"`
}

func (m *FinalRes) GetValid() bool {
	if m != nil {
		return m.Valid
	}
	return false
}

func (m *FinalRes) GetUserBlob() []byte {
	if m != nil {
		return m.UserBlob
	}
	return nil
}
