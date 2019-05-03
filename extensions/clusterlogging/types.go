package clusterlogging

type Project struct {
	Name string
	UUID string
}

type UserInfo struct {
	Username string
	Groups   []string
	Projects []Project
}
