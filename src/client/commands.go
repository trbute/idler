package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (m *uiModel) createCharacter(charName string) tea.Cmd {
	return func() tea.Msg {
		data := map[string]string{
			"name": charName,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		var bodyStr string
		var resColor Color
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = "Character Creation Successful"
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}


func (m *uiModel) getArea() tea.Cmd {
	return func() tea.Msg {
		bodyStr := ""
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected"}
		}

		req, err := http.NewRequest(
			"GET",
			m.apiUrl+fmt.Sprintf("/sense/area/%v", m.selectedChar),
			nil,
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			caser := cases.Title(language.English)
			var res senseAreaResponse
			if err := json.Unmarshal(body, &res); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = "\n"
				if len(res.Characters) > 0 {
					bodyStr += "Characters\n"
					for _, value := range res.Characters {
						if value.ActionName == "IDLE" || value.ActionTarget == "" {
							bodyStr += fmt.Sprintf(
								"\t%v is idle\n",
								value.CharacterName,
							)
						} else {
							bodyStr += fmt.Sprintf(
								"\t%v is %v at %v\n",
								value.CharacterName,
								caser.String(value.ActionName),
								caser.String(value.ActionTarget),
							)
						}
					}
				}

				if len(res.ResourceNodes) > 0 {
					bodyStr += "Resources\n"
					for _, value := range res.ResourceNodes {
						resource := caser.String(value)
						bodyStr += fmt.Sprintf("\t%v\n", resource)
					}
				}
			}
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) getInventory() tea.Cmd {
	return func() tea.Msg {
		bodyStr := ""
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected"}
		}

		req, err := http.NewRequest(
			"GET",
			m.apiUrl+fmt.Sprintf("/inventory/%v", m.selectedChar),
			nil,
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			caser := cases.Title(language.English)
			var res inventoryResponse
			if err := json.Unmarshal(body, &res); err != nil {
				bodyStr = err.Error()
			} else {
				bodyStr = "\n"
				if len(res.Items) > 0 {
					bodyStr += "Inventory\n"
					for name, item := range res.Items {
						bodyStr += fmt.Sprintf(
							"\t%v: %v (weight: %d each, total: %d)\n",
							caser.String(name),
							item.Quantity,
							item.Weight,
							item.TotalWeight,
						)
					}
				}
				bodyStr += fmt.Sprintf("\nWeight: %d/%d", res.Weight, res.Capacity)
			}
		} else {
			resColor = Red
			bodyStr = fmt.Sprintf("Inventory get failed for %v", m.selectedChar)
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) setIdle() tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		data := map[string]string{
			"character_name": m.selectedChar,
			"target":         "IDLE",
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"PUT",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 201 {
			resColor = Green
			bodyStr = fmt.Sprintf("%v is now idle", m.selectedChar)
		} else {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) dropItem(itemName, quantityStr string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		quantity, err := strconv.Atoi(quantityStr)
		if err != nil || quantity <= 0 {
			return apiResMsg{Red, "Invalid quantity. Must be a positive number"}
		}

		data := map[string]interface{}{
			"character_name": m.selectedChar,
			"item_name":      itemName,
			"quantity":       quantity,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/inventory/drop",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 200 {
			resColor = Green
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				if message, ok := response["message"].(string); ok {
					bodyStr = message
				} else {
					bodyStr = "Item dropped successfully"
				}
			} else {
				bodyStr = "Item dropped successfully"
			}
		} else {
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) dropItemAll(itemName string) tea.Cmd {
	return func() tea.Msg {
		if m.selectedChar == "" {
			return apiResMsg{Red, "No character selected. Use 'sel <character>' first"}
		}

		data := map[string]interface{}{
			"character_name": m.selectedChar,
			"item_name":      itemName,
			"drop_all":       true,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"POST",
			m.apiUrl+"/inventory/drop",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		var bodyStr string
		var resColor Color
		if res.StatusCode == 200 {
			resColor = Green
			var response map[string]interface{}
			if err := json.Unmarshal(body, &response); err == nil {
				if msg, ok := response["message"].(string); ok {
					bodyStr = msg
				} else {
					bodyStr = "All items dropped successfully"
				}
			} else {
				bodyStr = "All items dropped successfully"
			}
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) setActionWithAmount(target string, amount *int) tea.Cmd {
	return func() tea.Msg {
		data := map[string]interface{}{
			"character_name": m.selectedChar,
			"target":         target,
		}
		
		if amount != nil {
			data["amount"] = *amount
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req, err := http.NewRequest(
			"PUT",
			m.apiUrl+"/characters",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		bodyStr := ""
		resColor := Red
		if res.StatusCode == 201 {
			caser := cases.Title(language.English)
			resColor = Green

			var response map[string]interface{}
			json.Unmarshal(body, &response)
			actionName := response["action_name"].(string)

			bodyStr = fmt.Sprintf(
				"%v started %v on %v",
				caser.String(m.selectedChar),
				caser.String(actionName),
				caser.String(target),
			)

		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}

func (m *uiModel) selectCharacter(charName string) tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequest(
			"GET",
			m.apiUrl+fmt.Sprintf("/characters/%s/select", charName),
			nil,
		)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		req.Header.Set("Authorization", "Bearer "+m.userToken)

		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return apiResMsg{Red, err.Error()}
		}

		var bodyStr string
		var resColor Color
		if res.StatusCode == 200 {
			resColor = Green
			m.selectedChar = charName
			bodyStr = fmt.Sprintf("Selected %v", charName)
		} else {
			resColor = Red
			var errResp ErrorResponse
			if err := json.Unmarshal(body, &errResp); err != nil {
				bodyStr = "Failed to parse error response"
			} else {
				bodyStr = errResp.Error
			}
		}

		return apiResMsg{resColor, bodyStr}
	}
}
