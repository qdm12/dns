package splash

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyokomi/emoji"
	"github.com/qdm12/dns/internal/models"
)

// Splash returns the welcome spash message.
func Splash(buildInfo models.BuildInformation) string {
	lines := title()
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(
		"Running version %s built on %s (commit %s)",
		buildInfo.Version, buildInfo.BuildDate, buildInfo.Commit))
	lines = append(lines, "")
	lines = append(lines, printAnnouncement()...)
	lines = append(lines, "")
	lines = append(lines, links()...)
	return strings.Join(lines, "\n")
}

func title() []string {
	return []string{
		"=========================================",
		"========= DNS over TLS container ========",
		"=========================================",
		"=========================================",
		"=== Made with " + emoji.Sprint(":heart:") + " by github.com/qdm12 ====",
		"=========================================",
	}
}

func printAnnouncement() []string {
	if len(announcement) == 0 {
		return nil
	}
	expirationDate, _ := time.Parse("2006-01-02", announcementExpiration) // error covered by a unit test
	if time.Now().After(expirationDate) {
		return nil
	}
	return []string{emoji.Sprint(":mega: ") + announcement}
}

func links() []string {
	return []string{
		emoji.Sprint(":wrench: ") + "Need help? " + issueLink,
		emoji.Sprint(":computer: ") + "Email? quentin.mcgaw@gmail.com",
		emoji.Sprint(":coffee: ") + "Slack? Join from the Slack button on Github",
		emoji.Sprint(":money_with_wings: ") + "Help me? https://github.com/sponsors/qdm12",
	}
}
