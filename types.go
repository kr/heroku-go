package heroku

type App struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Maintenance bool   `json:"maintenance"`
}

func (a *App) Path() string {
	if a.ID == "" {
		return "/apps/" + a.Name
	}
	return "/apps/" + a.ID
}
