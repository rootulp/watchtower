package notifications

import (
	ty "github.com/containrrr/watchtower/pkg/types"
	"github.com/johntdyer/slackrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// NewNotifier creates and returns a new Notifier, using global configuration.
func NewNotifier(c *cobra.Command) ty.Notifier {
	f := c.PersistentFlags()

	level, _ := f.GetString("notifications-level")
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Fatalf("Notifications invalid log level: %s", err.Error())
	}

	acceptedLogLevels := slackrus.LevelThreshold(logLevel)
	// slackrus does not allow log level TRACE, even though it's an accepted log level for logrus
	if len(acceptedLogLevels) == 0 {
		log.Fatalf("Unsupported notification log level provided: %s", level)
	}

	reportTemplate, _ := f.GetBool("notification-report")
	tplString, _ := f.GetString("notification-template")
	urls, _ := f.GetStringArray("notification-url")

	urls = AppendLegacyUrls(urls, c)

	title := GetTitle(c)
	return newShoutrrrNotifier(tplString, acceptedLogLevels, !reportTemplate, title, urls...)
}

// AppendLegacyUrls creates shoutrrr equivalent URLs from legacy notification flags
func AppendLegacyUrls(urls []string, cmd *cobra.Command) []string {

	// Parse types and create notifiers.
	types, err := cmd.Flags().GetStringSlice("notifications")
	if err != nil {
		log.WithError(err).Fatal("could not read notifications argument")
	}

	for _, t := range types {

		var legacyNotifier ty.ConvertibleNotifier
		var err error

		switch t {
		case emailType:
			legacyNotifier = newEmailNotifier(cmd, []log.Level{})
		case slackType:
			legacyNotifier = newSlackNotifier(cmd, []log.Level{})
		case msTeamsType:
			legacyNotifier = newMsTeamsNotifier(cmd, []log.Level{})
		case gotifyType:
			legacyNotifier = newGotifyNotifier(cmd, []log.Level{})
		case shoutrrrType:
			continue
		default:
			log.Fatalf("Unknown notification type %q", t)
			// Not really needed, used for nil checking static analysis
			continue
		}

		shoutrrrURL, err := legacyNotifier.GetURL(cmd)
		if err != nil {
			log.Fatal("failed to create notification config:", err)
		}
		urls = append(urls, shoutrrrURL)

		log.WithField("URL", shoutrrrURL).Trace("created Shoutrrr URL from legacy notifier")
	}
	return urls
}

// GetTitle returns a common notification title with hostname appended
func GetTitle(c *cobra.Command) (title string) {
	title = "Watchtower updates"

	f := c.PersistentFlags()

	hostname, _ := f.GetString("notifications-hostname")

	if hostname != "" {
		title += " on " + hostname
	} else if hostname, err := os.Hostname(); err == nil {
		title += " on " + hostname
	}

	return
}

// ColorHex is the default notification color used for services that support it (formatted as a CSS hex string)
const ColorHex = "#406170"

// ColorInt is the default notification color used for services that support it (as an int value)
const ColorInt = 0x406170
