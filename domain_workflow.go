package main

type PlayerAsset struct{ PlayerID, AssetID, Kind string }
type LiveEvent struct{ ID, Name string }
type ModerationItem struct{ AssetID, Status string }

type OnboardingResult struct{ Domain, ZoneID, WebhookID, State string }

func onboard(c *Client, domainName, webhookURL string) (OnboardingResult, error) {
	d, err := c.AddDomain(domainName)
	if err != nil {
		return OnboardingResult{}, err
	}
	if _, err := c.UpsertRecord(d.ZoneID, "CNAME", "play", "edge.example.net"); err != nil {
		return OnboardingResult{}, err
	}
	if err := c.VerifyDomain(domainName); err != nil {
		return OnboardingResult{}, err
	}
	w, err := c.RegisterWebhook(webhookURL, []string{"dns.domain.verified"})
	if err != nil {
		return OnboardingResult{}, err
	}
	return OnboardingResult{Domain: domainName, ZoneID: d.ZoneID, WebhookID: w.ID, State: "active"}, nil
}
