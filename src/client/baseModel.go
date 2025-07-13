package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	tea "github.com/charmbracelet/bubbletea"
)

type Page int

const (
	Login Page = iota
	Signup
	UI
)

type apiResMsg struct {
	color Color
	text  string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *sharedState) clearSession() {
	s.currentPage = Login
	s.userToken = ""
	s.refreshToken = ""
}

type sharedState struct {
	*style
	currentPage    Page
	userToken      string
	refreshToken   string
	surname        string
	apiUrl         string
	wsUrl          string
	logoutMessage  string
	logoutMsgColor Color
}

type baseModel struct {
	*sharedState
	login  *loginModel
	signup *signupModel
	ui     *uiModel
}

func (m baseModel) Init() tea.Cmd {
	return nil
}

func (m baseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}


	switch m.currentPage {
	case Login:
		cmd = m.login.Update(msg)
	case Signup:
		cmd = m.signup.Update(msg)
	case UI:
		cmd = m.ui.Update(msg)
	}

	return m, cmd
}

func (m baseModel) View() string {
	var s string
	switch m.currentPage {
	case Login:
		s = m.login.View()
	case Signup:
		s = m.signup.View()
	case UI:
		s = m.ui.View()
	default:
		s = "Invalid View"
	}

	return s
}

type refreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *sharedState) refreshAuthToken() error {
	if s.refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	req, err := http.NewRequest("POST", s.apiUrl+"/refresh", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.refreshToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("refresh failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var refreshResp refreshTokenResponse
	err = json.Unmarshal(body, &refreshResp)
	if err != nil {
		return err
	}

	s.userToken = refreshResp.Token
	s.refreshToken = refreshResp.RefreshToken
	return nil
}

func (s *sharedState) makeAuthenticatedRequest(method, endpoint string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, s.apiUrl+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.userToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// If 401, try to refresh token and retry once
	if resp.StatusCode == 401 {
		resp.Body.Close()
		
		err = s.refreshAuthToken()
		if err != nil {
			s.clearSession()
			return nil, fmt.Errorf("session_expired")
		}

		req.Header.Set("Authorization", "Bearer "+s.userToken)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		
		if resp.StatusCode == 401 {
			resp.Body.Close()
			s.clearSession()
			return nil, fmt.Errorf("session_expired")
		}
	}

	return resp, nil
}

func (s *sharedState) handleAPIError(err error) apiResMsg {
	if err.Error() == "session_expired" {
		return apiResMsg{Yellow, "Session expired. Please login again."}
	}
	return apiResMsg{Red, err.Error()}
}
