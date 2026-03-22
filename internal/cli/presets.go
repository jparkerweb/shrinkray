package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jparkerweb/shrinkray/internal/presets"
)

func presetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "presets",
		Short: "List all available encoding presets",
		RunE:  listPresets,
	}

	cmd.AddCommand(presetsShowCmd())

	return cmd
}

func listPresets(cmd *cobra.Command, args []string) error {
	allPresets := presets.All()

	// Header
	fmt.Printf("%-16s %-22s %-10s %-6s %-6s %-10s\n",
		"KEY", "NAME", "CATEGORY", "CODEC", "CRF", "TARGET")
	fmt.Println(strings.Repeat("-", 74))

	// Group by category
	categoryOrder := []presets.Category{
		presets.CategoryQuality,
		presets.CategoryPurpose,
		presets.CategoryPlatform,
	}

	for _, cat := range categoryOrder {
		for _, p := range allPresets {
			if p.Category != cat {
				continue
			}

			target := ""
			if p.TargetSizeMB > 0 {
				target = fmt.Sprintf("%.0f MB", p.TargetSizeMB)
			}

			crfStr := fmt.Sprintf("%d", p.CRF)
			if p.CRF == 0 && p.TargetSizeMB > 0 {
				crfStr = "-"
			}

			fmt.Printf("%-16s %-22s %-10s %-6s %-6s %-10s\n",
				p.Key,
				truncate(p.Name, 22),
				p.Category,
				strings.ToUpper(p.Codec),
				crfStr,
				target,
			)
		}
	}

	fmt.Printf("\n%d presets available. Use 'shrinkray presets show <key>' for details.\n", len(allPresets))

	return nil
}

func presetsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [key]",
		Short: "Show detailed information about a preset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			p, found := presets.Lookup(key)
			if !found {
				return fmt.Errorf("preset not found: %s\nRun 'shrinkray presets' to see available presets", key)
			}

			fmt.Printf("%s %s\n", p.Icon, p.Name)
			fmt.Printf("  Key:         %s\n", p.Key)
			fmt.Printf("  Category:    %s\n", p.Category)
			fmt.Printf("  Description: %s\n", p.Description)
			fmt.Printf("\n")
			fmt.Printf("  Codec:       %s\n", strings.ToUpper(p.Codec))
			fmt.Printf("  Container:   %s\n", p.Container)
			fmt.Printf("  CRF:         %d\n", p.CRF)
			if p.MaxBitrate != "" {
				fmt.Printf("  Max Bitrate: %s\n", p.MaxBitrate)
			}
			fmt.Printf("  Speed:       %s\n", p.SpeedPreset)
			if p.Resolution != "" {
				fmt.Printf("  Resolution:  %s\n", p.Resolution)
			}
			if p.MaxFPS > 0 {
				fmt.Printf("  Max FPS:     %d\n", p.MaxFPS)
			}
			fmt.Printf("\n")
			if p.AudioCodec == "copy" {
				fmt.Printf("  Audio:       copy (passthrough)\n")
			} else {
				fmt.Printf("  Audio:       %s", p.AudioCodec)
				if p.AudioBitrate != "" {
					fmt.Printf(" @ %s", p.AudioBitrate)
				}
				fmt.Printf("\n")
			}
			if p.AudioChannels > 0 {
				fmt.Printf("  Channels:    %d\n", p.AudioChannels)
			}
			fmt.Printf("\n")
			if p.TargetSizeMB > 0 {
				fmt.Printf("  Target Size: %.0f MB\n", p.TargetSizeMB)
			}
			if p.TwoPass {
				fmt.Printf("  Encoding:    Two-pass\n")
			}
			if len(p.Tags) > 0 {
				fmt.Printf("  Tags:        %s\n", strings.Join(p.Tags, ", "))
			}

			return nil
		},
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
