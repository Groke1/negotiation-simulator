package organizations

type createOrganizationRequest struct {
	Name string `json:"name"`
}

type organizationResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type organizationListResponse struct {
	Organizations []organizationResponse `json:"organizations"`
}
