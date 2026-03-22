package cli

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/jparkerweb/shrinkray/internal/config"
	"github.com/jparkerweb/shrinkray/internal/engine"
	"github.com/jparkerweb/shrinkray/internal/logging"
	"github.com/jparkerweb/shrinkray/internal/presets"
	"github.com/jparkerweb/shrinkray/internal/tui"
	"github.com/jparkerweb/shrinkray/internal/tui/screens"
)

// Run flags
var (
	flagInput      string
	flagInputs     []string // multiple -i flags
	flagPreset     string
	flagOutput     string
	flagNoTUI      bool
	flagCRF        int
	flagCodec      string
	flagResolution string
	flagSuffix     string

	// Batch flags
	flagJobs         int
	flagRecursive    bool
	flagSort         string
	flagSkipExisting bool
	flagSkipOptimal  bool
	flagInPlace      bool
	flagOutputDir    string
	flagRetryFailed  bool
	flagMaxRetries   int

	// Phase 6 flags
	flagDryRun        bool
	flagStdin         bool
	flagStripMetadata bool
	flagKeepMetadata  bool
	flagMetadataTitle string
	flagOpen          bool
)

func registerRunFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&flagInput, "input", "i", "", "input video file path")
	cmd.Flags().StringArrayVar(&flagInputs, "inputs", nil, "additional input video file paths")
	cmd.Flags().StringVarP(&flagPreset, "preset", "p", "", "preset name (e.g., balanced, compact, tiny)")
	cmd.Flags().StringVarP(&flagOutput, "output", "o", "", "output file path")
	cmd.Flags().BoolVar(&flagNoTUI, "no-tui", false, "disable interactive TUI, use headless mode")
	cmd.Flags().IntVar(&flagCRF, "crf", 0, "CRF value override (0 = use preset default)")
	cmd.Flags().StringVar(&flagCodec, "codec", "", "video codec (h264, h265, av1, vp9)")
	cmd.Flags().StringVar(&flagResolution, "resolution", "", "output resolution (e.g., 1920x1080)")
	cmd.Flags().StringVar(&flagSuffix, "suffix", "", "output filename suffix (default: _shrunk)")

	// Batch flags
	cmd.Flags().IntVarP(&flagJobs, "jobs", "j", 1, "number of parallel encoding workers")
	cmd.Flags().BoolVarP(&flagRecursive, "recursive", "r", false, "recurse into directories for video files")
	cmd.Flags().StringVar(&flagSort, "sort", "", "sort order: size-asc, size-desc, name, duration")
	cmd.Flags().BoolVar(&flagSkipExisting, "skip-existing", false, "skip files whose output already exists")
	cmd.Flags().BoolVar(&flagSkipOptimal, "skip-optimal", false, "skip files already compressed with target codec")
	cmd.Flags().BoolVar(&flagInPlace, "in-place", false, "replace source files after verification (destructive!)")
	cmd.Flags().StringVar(&flagOutputDir, "output-dir", "", "output directory (mirrors input structure)")
	cmd.Flags().BoolVar(&flagRetryFailed, "retry-failed", false, "retry failed jobs from persisted queue")
	cmd.Flags().IntVar(&flagMaxRetries, "max-retries", 2, "maximum retry attempts per file")

	// Dry-run, stdin, metadata, open flags
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "print FFmpeg command without executing")
	cmd.Flags().BoolVar(&flagStdin, "stdin", false, "read file paths from stdin (one per line)")
	cmd.Flags().BoolVar(&flagStripMetadata, "strip-metadata", false, "remove all metadata from output")
	cmd.Flags().BoolVar(&flagKeepMetadata, "keep-metadata", true, "preserve source metadata (default)")
	cmd.Flags().StringVar(&flagMetadataTitle, "metadata-title", "", "set output title metadata")
	cmd.Flags().BoolVar(&flagOpen, "open", false, "open output folder after completion")
}

// runCmd is the default command — handles both TUI and headless modes.
func runCmd(cmd *cobra.Command, args []string) error {
	// Determine if we should use TUI mode
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))

	// Stdin pipe input: if --stdin flag or stdin is not a terminal and no -i given
	if flagStdin || (!isTerminal && flagInput == "" && len(flagInputs) == 0) {
		stdinPaths, err := readStdinPaths()
		if err != nil {
			return err
		}
		if len(stdinPaths) == 0 {
			return fmt.Errorf("no valid file paths received from stdin")
		}
		// Force headless mode for stdin input
		flagNoTUI = true
		flagInputs = append(flagInputs, stdinPaths...)
	}

	if !flagNoTUI && isTerminal && !flagDryRun {
		return runTUI(cmd)
	}

	// Fall back to headless if --no-tui or non-TTY
	if flagInput == "" && len(flagInputs) == 0 {
		return fmt.Errorf("input file is required in headless mode (use -i flag)")
	}

	return runHeadless(cmd)
}

// readStdinPaths reads file paths from stdin, one per line.
// Trims whitespace, skips empty lines and lines starting with #.
// Validates each path exists and is a video file.
func readStdinPaths() ([]string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var paths []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Validate file exists
		stat, err := os.Stat(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", line, err)
			continue
		}
		if stat.IsDir() {
			found := screens.DiscoverVideoFiles(line, flagRecursive)
			paths = append(paths, found...)
			continue
		}
		if !isVideoFile(line) {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: not a video file\n", line)
			continue
		}
		paths = append(paths, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stdin: %w", err)
	}
	return paths, nil
}

func runTUI(cmd *cobra.Command) error {
	// Setup logging
	if err := logging.Setup("tui", flagLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to setup logging: %v\n", err)
	}

	// Load saved theme from config
	cfg, cfgErr := config.Load(flagConfig)
	if cfgErr == nil && cfg.UI.Theme != "" {
		tui.SetTheme(tui.ThemeName(cfg.UI.Theme))
	}

	opts := tui.AppOptions{}

	// If input file was provided via -i, probe it and pre-load
	if flagInput != "" {
		if _, err := os.Stat(flagInput); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("input file not found: %s", flagInput)
			}
			return fmt.Errorf("cannot access input file: %w", err)
		}

		opts.InputPath = flagInput

		// Probe the video
		videoInfo, err := engine.Probe(context.Background(), flagInput)
		if err != nil {
			return fmt.Errorf("failed to probe input: %w", err)
		}
		opts.VideoInfo = videoInfo
	}

	app := tui.NewApp(opts)
	p := tea.NewProgram(app)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func runHeadless(cmd *cobra.Command) error {
	// Setup logging
	if err := logging.Setup("headless", flagLogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to setup logging: %v\n", err)
	}

	// Load config
	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply CLI overrides
	overrides := config.ConfigOverrides{
		Preset: flagPreset,
		Codec:  flagCodec,
		Suffix: flagSuffix,
	}
	if cmd.Flags().Changed("crf") {
		overrides.CRF = &flagCRF
	}
	if flagOutputDir != "" {
		overrides.OutputDir = flagOutputDir
	}
	cfg.Merge(overrides)

	// Collect all input paths
	inputs := collectInputs()

	if len(inputs) == 0 {
		return fmt.Errorf("input file is required in headless mode (use -i flag)")
	}

	// Check for retry-failed mode
	if flagRetryFailed {
		return runRetryFailed(cmd, cfg)
	}

	// If multiple inputs or directory, run batch
	if len(inputs) > 1 {
		return runHeadlessBatch(cmd, cfg, inputs)
	}

	// Single file mode
	return runHeadlessSingle(cmd, cfg, inputs[0])
}

// collectInputs gathers all input paths from flags.
func collectInputs() []string {
	var inputs []string

	if flagInput != "" {
		// Check if it's a directory
		stat, err := os.Stat(flagInput)
		if err == nil && stat.IsDir() {
			found := screens.DiscoverVideoFiles(flagInput, flagRecursive)
			inputs = append(inputs, found...)
		} else {
			inputs = append(inputs, flagInput)
		}
	}

	for _, inp := range flagInputs {
		stat, err := os.Stat(inp)
		if err == nil && stat.IsDir() {
			found := screens.DiscoverVideoFiles(inp, flagRecursive)
			inputs = append(inputs, found...)
		} else {
			inputs = append(inputs, inp)
		}
	}

	return inputs
}

func runHeadlessSingle(cmd *cobra.Command, cfg *config.Config, input string) error {
	// Validate input
	if _, err := os.Stat(input); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", input)
		}
		return fmt.Errorf("cannot access input file: %w", err)
	}

	// Detect FFmpeg
	ffmpegInfo, err := engine.DetectFFmpeg()
	if err != nil {
		return err
	}
	slog.Info("ffmpeg detected", "path", ffmpegInfo.Path, "version", ffmpegInfo.Version)

	// Probe input video
	fmt.Fprintf(os.Stderr, "Probing %s...\n", input)
	videoInfo, err := engine.Probe(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to probe input: %w", err)
	}

	fmt.Fprintf(os.Stderr, "  Format:     %s\n", videoInfo.Format)
	fmt.Fprintf(os.Stderr, "  Resolution: %s\n", videoInfo.Resolution())
	fmt.Fprintf(os.Stderr, "  Codec:      %s\n", videoInfo.Codec)
	fmt.Fprintf(os.Stderr, "  Duration:   %s\n", formatDuration(videoInfo.Duration))
	fmt.Fprintf(os.Stderr, "  Size:       %s\n", formatBytes(videoInfo.Size))
	if videoInfo.IsHDR {
		fmt.Fprintf(os.Stderr, "  HDR:        %s\n", videoInfo.HDRFormat)
	}

	// Lookup preset
	presetName := cfg.Defaults.Preset
	preset, found := presets.Lookup(presetName)
	if !found {
		return fmt.Errorf("preset not found: %s\nRun 'shrinkray presets' to see available presets", presetName)
	}
	fmt.Fprintf(os.Stderr, "\nUsing preset: %s %s (%s)\n", preset.Icon, preset.Name, preset.Key)

	// Resolve output path
	outputOpts := buildOutputOpts(cfg)
	if flagOutput != "" {
		outputOpts.Mode = engine.OutputModeExplicit
		outputOpts.ExplicitPath = flagOutput
	}

	outputPath, err := engine.ResolveOutput(input, outputOpts)
	if err != nil {
		return fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Use temp file for safety
	tempPath := engine.TempPath(outputPath)
	defer func() { _ = os.Remove(tempPath) }()

	fmt.Fprintf(os.Stderr, "Output:       %s\n\n", outputPath)

	// Build encode options
	encodeOpts := engine.EncodeOptions{
		Input:  input,
		Output: tempPath,
		Preset: preset,
	}
	if cmd.Flags().Changed("crf") {
		encodeOpts.CRFOverride = &flagCRF
	}
	if flagResolution != "" {
		encodeOpts.ResolutionOverride = flagResolution
	}
	if flagCodec != "" {
		encodeOpts.Preset.Codec = flagCodec
	}

	// Metadata handling
	encodeOpts.MetadataMode = engine.MetadataFromFlags(flagStripMetadata, flagKeepMetadata, flagMetadataTitle)

	// Dry-run mode: print command and exit
	if flagDryRun {
		args := engine.BuildArgs(encodeOpts)
		fmt.Println("ffmpeg \\")
		for i, arg := range args {
			if i < len(args)-1 {
				fmt.Printf("  %s \\\n", arg)
			} else {
				fmt.Printf("  %s\n", arg)
			}
		}
		fmt.Printf("\n# Input:           %s (%s)\n", input, formatBytes(videoInfo.Size))
		fmt.Printf("# Output:          %s\n", outputPath)
		fmt.Printf("# Preset:          %s (%s)\n", preset.Name, preset.Key)

		estimated := engine.EstimateSize(videoInfo, preset)
		if estimated > 0 {
			savings := float64(videoInfo.Size-estimated) / float64(videoInfo.Size) * 100
			fmt.Printf("# Est. output:     %s (%.1f%% savings)\n", formatBytes(estimated), savings)
		}
		return nil
	}

	// Create cancellable context for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Fprintf(os.Stderr, "\nCancelling encode...\n")
		cancel()
	}()

	// Start encoding
	isTwoPass := engine.ShouldUseTwoPass(encodeOpts)
	if isTwoPass {
		if preset.TargetSizeMB > 0 {
			dur := time.Duration(videoInfo.Duration * float64(time.Second))
			audioBitrate := int64(128000)
			if videoInfo.AudioBitrate > 0 {
				audioBitrate = videoInfo.AudioBitrate
			}
			targetBytes := int64(preset.TargetSizeMB * 1024 * 1024)
			encodeOpts.VideoBitrate = engine.CalculateBitrate(targetBytes, dur, audioBitrate)

			adaptW, adaptH := engine.AdaptiveResolution(
				targetBytes, dur,
				videoInfo.Width, videoInfo.Height,
				videoInfo.Framerate,
			)
			if adaptW < videoInfo.Width || adaptH < videoInfo.Height {
				encodeOpts.ResolutionOverride = fmt.Sprintf("%dx%d", adaptW, adaptH)
				fmt.Fprintf(os.Stderr, "  Adaptive resolution: %dx%d (for target size)\n", adaptW, adaptH)
			}
		}
		fmt.Fprintf(os.Stderr, "Encoding (two-pass)...\n")
	} else {
		fmt.Fprintf(os.Stderr, "Encoding...\n")
	}
	startTime := time.Now()

	var progressCh <-chan engine.ProgressUpdate
	if isTwoPass {
		progressCh, err = engine.EncodeTwoPass(ctx, encodeOpts)
	} else {
		progressCh, err = engine.Encode(ctx, encodeOpts)
	}
	if err != nil {
		return fmt.Errorf("failed to start encoding: %w", err)
	}

	// Print progress updates
	var lastUpdate engine.ProgressUpdate
	for update := range progressCh {
		lastUpdate = update
		if update.Error != nil {
			return fmt.Errorf("encoding failed: %w", update.Error)
		}
		if !update.Done {
			fmt.Fprintf(os.Stderr, "\r  Progress: %5.1f%% | Speed: %.1fx | ETA: %s        ",
				update.Percent, update.Speed, formatDurationShort(update.ETA))
		}
	}

	if lastUpdate.Error != nil {
		return fmt.Errorf("encoding failed: %w", lastUpdate.Error)
	}

	elapsed := time.Since(startTime)

	// For in-place mode, verify before replacing
	if flagInPlace {
		if err := verifyAndReplace(ctx, input, tempPath); err != nil {
			return err
		}
		outputPath = input
	} else {
		// Move temp file to final output
		if err := os.Rename(tempPath, outputPath); err != nil {
			return fmt.Errorf("failed to move output file: %w", err)
		}
	}

	// Print summary
	outStat, err := os.Stat(outputPath)
	if err != nil {
		return fmt.Errorf("failed to stat output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\r  Progress: 100.0%% | Done!                              \n\n")

	inputSize := videoInfo.Size
	outputSize := outStat.Size()
	savings := float64(inputSize-outputSize) / float64(inputSize) * 100
	ratio := float64(inputSize) / float64(outputSize)

	fmt.Printf("Encoding complete!\n")
	fmt.Printf("  Input:       %s (%s)\n", input, formatBytes(inputSize))
	fmt.Printf("  Output:      %s (%s)\n", outputPath, formatBytes(outputSize))
	fmt.Printf("  Savings:     %.1f%% (%.1fx compression)\n", savings, ratio)
	fmt.Printf("  Time:        %s\n", formatDurationShort(elapsed))

	// Open folder if requested
	if flagOpen {
		if err := engine.OpenFolder(outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not open folder: %v\n", err)
		}
	}

	return nil
}

func runHeadlessBatch(cmd *cobra.Command, cfg *config.Config, inputs []string) error {
	// Detect FFmpeg
	ffmpegInfo, err := engine.DetectFFmpeg()
	if err != nil {
		return err
	}
	slog.Info("ffmpeg detected", "path", ffmpegInfo.Path, "version", ffmpegInfo.Version)

	// Lookup preset
	presetName := cfg.Defaults.Preset
	preset, found := presets.Lookup(presetName)
	if !found {
		return fmt.Errorf("preset not found: %s\nRun 'shrinkray presets' to see available presets", presetName)
	}

	fmt.Fprintf(os.Stderr, "Batch encoding %d files with preset: %s %s\n", len(inputs), preset.Icon, preset.Name)

	// Build job queue
	queue := engine.NewJobQueue()
	for _, input := range inputs {
		stat, err := os.Stat(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: skipping %s: %v\n", input, err)
			continue
		}
		var duration float64
		info, probeErr := engine.Probe(context.Background(), input)
		if probeErr == nil {
			duration = info.Duration
		}
		queue.Add(engine.Job{
			InputPath: input,
			PresetKey: preset.Key,
			Status:    engine.JobStatusPending,
			InputSize: stat.Size(),
			Duration:  duration,
		})
	}

	// Sort if requested
	sortMode := flagSort
	if sortMode == "" {
		sortMode = cfg.Batch.Sort
	}
	if sortMode != "" {
		queue.SortQueue(sortMode)
	}

	// Dry-run for batch mode
	if flagDryRun {
		outputOpts := buildOutputOpts(cfg)
		for _, job := range queue.All() {
			out, err := engine.ResolveOutput(job.InputPath, outputOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "# Cannot resolve output for %s: %v\n", job.InputPath, err)
				continue
			}
			encOpts := engine.EncodeOptions{
				Input:  job.InputPath,
				Output: out,
				Preset: preset,
			}
			encOpts.MetadataMode = engine.MetadataFromFlags(flagStripMetadata, flagKeepMetadata, flagMetadataTitle)
			args := engine.BuildArgs(encOpts)
			fmt.Printf("# File: %s\n", job.InputPath)
			fmt.Println("ffmpeg \\")
			for i, arg := range args {
				if i < len(args)-1 {
					fmt.Printf("  %s \\\n", arg)
				} else {
					fmt.Printf("  %s\n", arg)
				}
			}
			fmt.Println()
		}
		return nil
	}

	// Build output options
	outputOpts := buildOutputOpts(cfg)

	// Set base dir for directory mirroring
	if outputOpts.Mode == engine.OutputModeDirectory && flagInput != "" {
		stat, err := os.Stat(flagInput)
		if err == nil && stat.IsDir() {
			outputOpts.BaseDir = flagInput
		}
	}

	// Build skip options
	skipOpts := engine.SkipOptions{
		SkipExisting: flagSkipExisting || cfg.Batch.SkipExisting,
		SkipOptimal:  flagSkipOptimal,
	}

	// Setup persistence
	queuePath, _ := engine.QueuePath()
	saver := engine.NewDebouncedSaver(queuePath, queue)

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Fprintf(os.Stderr, "\nCancelling batch...\n")
		cancel()
		saver.Flush()
	}()

	jobs := flagJobs
	if jobs < 1 {
		jobs = cfg.Batch.Jobs
	}
	if jobs < 1 {
		jobs = 1
	}

	maxRetries := flagMaxRetries
	if maxRetries < 1 {
		maxRetries = 2
	}

	batchOpts := engine.BatchOptions{
		Jobs:       jobs,
		Preset:     preset,
		OutputOpts: outputOpts,
		SkipOpts:   skipOpts,
		MaxRetries: maxRetries,
		OnSave:     saver.Save,
	}

	startTime := time.Now()
	eventCh := engine.RunBatch(ctx, queue, batchOpts)

	total := queue.Len()
	processed := 0

	for event := range eventCh {
		switch e := event.(type) {
		case engine.JobStartedEvent:
			processed++
			fmt.Fprintf(os.Stderr, "[%d/%d] encoding %s...\n", processed, total, filepath.Base(e.InputPath))

		case engine.JobProgressEvent:
			fmt.Fprintf(os.Stderr, "\r  [%d/%d] %s ... %.0f%% %.1fx ETA %s        ",
				processed, total,
				filepath.Base(e.JobID), // We need the path, not ID
				e.Progress, e.Update.Speed,
				formatDurationShort(e.Update.ETA))

		case engine.JobCompleteEvent:
			fmt.Fprintf(os.Stderr, "\r  [%d/%d] %s ... done (%s -> %s)                    \n",
				processed, total,
				filepath.Base(e.InputPath),
				formatBytes(e.InputSize),
				formatBytes(e.OutputSize))

		case engine.JobFailedEvent:
			fmt.Fprintf(os.Stderr, "\r  [%d/%d] %s ... FAILED: %s                    \n",
				processed, total,
				filepath.Base(e.InputPath), e.Error)

		case engine.JobSkippedEvent:
			processed++
			fmt.Fprintf(os.Stderr, "  [%d/%d] %s ... skipped: %s\n",
				processed, total,
				filepath.Base(e.InputPath), e.Reason)

		case engine.BatchCompleteEvent:
			elapsed := time.Since(startTime)
			saver.Flush()

			stats := e.Stats
			fmt.Fprintf(os.Stderr, "\n")
			fmt.Printf("Batch encoding complete!\n")
			fmt.Printf("  Files:       %d total (%d complete, %d failed, %d skipped)\n",
				stats.Total, stats.Complete, stats.Failed, stats.Skipped)
			fmt.Printf("  Input:       %s\n", formatBytes(stats.TotalInputSize))
			fmt.Printf("  Output:      %s\n", formatBytes(stats.TotalOutputSize))
			if stats.TotalInputSize > 0 && stats.TotalOutputSize > 0 {
				savings := float64(stats.TotalInputSize-stats.TotalOutputSize) / float64(stats.TotalInputSize) * 100
				fmt.Printf("  Savings:     %.1f%%\n", savings)
			}
			fmt.Printf("  Time:        %s\n", formatDurationShort(elapsed))

			// Clean queue if all complete
			if stats.Failed == 0 && stats.Pending == 0 {
				_ = engine.CleanQueue("")
			}
		}
	}

	return nil
}

func runRetryFailed(cmd *cobra.Command, cfg *config.Config) error {
	hasPending, queue := engine.HasPendingQueue("")
	if !hasPending || queue == nil {
		return fmt.Errorf("no saved queue with failed/pending jobs found")
	}

	stats := queue.Stats()
	fmt.Fprintf(os.Stderr, "Found saved queue: %d pending, %d failed\n", stats.Pending, stats.Failed)

	// Re-queue failed jobs as pending
	failed := queue.ByStatus(engine.JobStatusFailed)
	for _, job := range failed {
		queue.Update(job.ID, func(j *engine.Job) {
			j.Status = engine.JobStatusPending
			j.Progress = 0
			j.Error = ""
		})
	}

	// Lookup preset from first job
	presetKey := cfg.Defaults.Preset
	if len(queue.All()) > 0 {
		first := queue.All()[0]
		if first.PresetKey != "" {
			presetKey = first.PresetKey
		}
	}

	preset, found := presets.Lookup(presetKey)
	if !found {
		return fmt.Errorf("preset not found: %s", presetKey)
	}

	// Re-run as batch
	inputs := make([]string, 0)
	for _, j := range queue.ByStatus(engine.JobStatusPending) {
		inputs = append(inputs, j.InputPath)
	}

	fmt.Fprintf(os.Stderr, "Retrying %d files with preset: %s\n", len(inputs), preset.Name)

	// Build the same batch options
	outputOpts := buildOutputOpts(cfg)
	skipOpts := engine.SkipOptions{
		SkipExisting: flagSkipExisting,
		SkipOptimal:  flagSkipOptimal,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		cancel()
	}()

	batchOpts := engine.BatchOptions{
		Jobs:       flagJobs,
		Preset:     preset,
		OutputOpts: outputOpts,
		SkipOpts:   skipOpts,
		MaxRetries: flagMaxRetries,
	}

	if batchOpts.Jobs < 1 {
		batchOpts.Jobs = 1
	}

	startTime := time.Now()
	eventCh := engine.RunBatch(ctx, queue, batchOpts)
	total := len(inputs)
	processed := 0

	for event := range eventCh {
		switch e := event.(type) {
		case engine.JobStartedEvent:
			processed++
			fmt.Fprintf(os.Stderr, "[%d/%d] retrying %s...\n", processed, total, filepath.Base(e.InputPath))
		case engine.JobCompleteEvent:
			fmt.Fprintf(os.Stderr, "  done: %s -> %s\n", formatBytes(e.InputSize), formatBytes(e.OutputSize))
		case engine.JobFailedEvent:
			fmt.Fprintf(os.Stderr, "  FAILED: %s\n", e.Error)
		case engine.BatchCompleteEvent:
			elapsed := time.Since(startTime)
			fmt.Printf("Retry complete in %s: %d complete, %d failed\n",
				formatDurationShort(elapsed), e.Stats.Complete, e.Stats.Failed)
		default:
		}
	}

	return nil
}

func buildOutputOpts(cfg *config.Config) engine.OutputOptions {
	opts := engine.OutputOptions{
		Suffix:       cfg.Output.Suffix,
		ConflictMode: engine.ConflictMode(cfg.Output.Conflict),
	}

	if flagInPlace {
		opts.Mode = engine.OutputModeInplace
	} else if flagOutputDir != "" {
		opts.Mode = engine.OutputModeDirectory
		opts.Directory = flagOutputDir
	} else {
		opts.Mode = engine.OutputMode(cfg.Output.Mode)
	}

	if flagSkipExisting {
		opts.ConflictMode = engine.ConflictSkip
	}

	return opts
}

func verifyAndReplace(ctx context.Context, sourcePath, tempPath string) error {
	sourceInfo, err := engine.Probe(ctx, sourcePath)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("in-place verification failed: cannot probe source: %w", err)
	}

	tempInfo, err := engine.Probe(ctx, tempPath)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("in-place verification failed: output is not a valid video: %w", err)
	}

	if tempInfo.Width <= 0 || tempInfo.Height <= 0 {
		os.Remove(tempPath)
		return fmt.Errorf("in-place verification failed: output has no video stream")
	}

	if sourceInfo.Duration > 0 {
		diff := tempInfo.Duration - sourceInfo.Duration
		if diff < 0 {
			diff = -diff
		}
		tolerance := sourceInfo.Duration * 0.05
		if diff > tolerance {
			os.Remove(tempPath)
			return fmt.Errorf("in-place verification failed: output duration %.1fs differs from source %.1fs by more than 5%%",
				tempInfo.Duration, sourceInfo.Duration)
		}
	}

	// Replace source with temp
	if err := os.Rename(tempPath, sourcePath); err != nil {
		return fmt.Errorf("in-place replace failed: %w", err)
	}

	return nil
}

func formatDuration(seconds float64) string {
	d := time.Duration(seconds * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatDurationShort(d time.Duration) string {
	if d <= 0 {
		return "--:--"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%02ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func formatBytes(b int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// isVideoFile checks if a file path has a supported video extension.
func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	videoExts := []string{".mp4", ".mkv", ".webm", ".avi", ".mov", ".wmv", ".flv", ".ts", ".m4v"}
	for _, ve := range videoExts {
		if ext == ve {
			return true
		}
	}
	return false
}
