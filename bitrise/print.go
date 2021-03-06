package bitrise

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/bitrise-io/bitrise/models"
	"github.com/bitrise-io/go-utils/colorstring"
	"github.com/bitrise-io/go-utils/stringutil"
	"github.com/bitrise-io/go-utils/versions"
	stepmanModels "github.com/bitrise-io/stepman/models"
)

const (
	// should not be under ~45
	stepRunSummaryBoxWidthInChars = 80
)

//------------------------------
// Util methods
//------------------------------

func isUpdateAvailable(stepInfo stepmanModels.StepInfoModel) bool {
	if stepInfo.Latest == "" {
		return false
	}

	res, err := versions.CompareVersions(stepInfo.Version, stepInfo.Latest)
	if err != nil {
		log.Debugf("Failed to compare versions, err: %s", err)
	}

	return (res == 1)
}

func getTrimmedStepName(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	stepInfo := stepRunResult.StepInfo

	title := stepInfo.Title
	if stepInfo.GlobalInfo.RemovalDate != "" {
		title = fmt.Sprintf("[Deprecated] %s", title)
	}

	titleBox := ""
	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess, models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		titleBox = fmt.Sprintf("%s", title)
		if len(titleBox) > titleBoxWidth {
			dif := len(titleBox) - titleBoxWidth
			title = stringutil.MaxFirstCharsWithDots(title, len(title)-dif)
			titleBox = fmt.Sprintf("%s", title)
		}
		break
	case models.StepRunStatusCodeFailed, models.StepRunStatusCodeFailedSkippable:
		titleBox = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		if len(titleBox) > titleBoxWidth {
			dif := len(titleBox) - titleBoxWidth
			title = stringutil.MaxFirstCharsWithDots(title, len(title)-dif)
			titleBox = fmt.Sprintf("%s (exit code: %d)", title, stepRunResult.ExitCode)
		}
		break
	default:
		log.Error("Unkown result code")
		return ""
	}

	return titleBox
}

func getRunningStepHeaderMainSection(stepInfo stepmanModels.StepInfoModel, idx int) string {
	title := stepInfo.Title

	content := fmt.Sprintf("| (%d) %s |", idx, title)
	charDiff := len(content) - stepRunSummaryBoxWidthInChars

	if charDiff < 0 {
		// shorter than desired - fill with space
		content = fmt.Sprintf("| (%d) %s%s |", idx, title, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedTitleWidth := len(title) - charDiff
		if trimmedTitleWidth < 4 {
			log.Errorf("Step title too long, can't present title at all! : %s", title)
		} else {
			content = fmt.Sprintf("| (%d) %s |", idx, stringutil.MaxFirstCharsWithDots(title, trimmedTitleWidth))
		}
	}
	return content
}

func getRunningStepHeaderSubSection(stepInfo stepmanModels.StepInfoModel) string {
	id := stepInfo.ID
	version := stepInfo.Version
	collection := stepInfo.StepLib
	logTime := time.Now().Format(time.RFC3339)

	idRow := fmt.Sprintf("| id: %s |", id)
	charDiff := len(idRow) - stepRunSummaryBoxWidthInChars
	if charDiff < 0 {
		// shorter than desired - fill with space
		idRow = fmt.Sprintf("| id: %s%s |", id, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedWidth := len(id) - charDiff
		if trimmedWidth < 4 {
			log.Errorf("Step id too long, can't present id at all! : %s", id)
		} else {
			idRow = fmt.Sprintf("| id: %s |", stringutil.MaxFirstCharsWithDots(id, trimmedWidth))
		}
	}

	versionRow := fmt.Sprintf("| version: %s |", version)
	charDiff = len(versionRow) - stepRunSummaryBoxWidthInChars
	if charDiff < 0 {
		// shorter than desired - fill with space
		versionRow = fmt.Sprintf("| version: %s%s |", version, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedWidth := len(version) - charDiff
		if trimmedWidth < 4 {
			log.Errorf("Step version too long, can't present version at all! : %s", version)
		} else {
			versionRow = fmt.Sprintf("| id: %s |", stringutil.MaxFirstCharsWithDots(version, trimmedWidth))
		}
	}

	collectionRow := fmt.Sprintf("| collection: %s |", collection)
	charDiff = len(collectionRow) - stepRunSummaryBoxWidthInChars
	if charDiff < 0 {
		// shorter than desired - fill with space
		collectionRow = fmt.Sprintf("| collection: %s%s |", collection, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedWidth := len(collection) - charDiff
		if trimmedWidth < 4 {
			log.Errorf("Step collection too long, can't present collection at all! : %s", version)
		} else {
			collectionRow = fmt.Sprintf("| collection: %s |", stringutil.MaxLastCharsWithDots(collection, trimmedWidth))
		}
	}

	timeRow := fmt.Sprintf("| time: %s |", logTime)
	charDiff = len(timeRow) - stepRunSummaryBoxWidthInChars
	if charDiff < 0 {
		// shorter than desired - fill with space
		timeRow = fmt.Sprintf("| time: %s%s |", logTime, strings.Repeat(" ", -charDiff))
	} else if charDiff > 0 {
		// longer than desired - trim title
		trimmedWidth := len(logTime) - charDiff
		if trimmedWidth < 4 {
			log.Errorf("Time too long, can't present time at all! : %s", version)
		} else {
			timeRow = fmt.Sprintf("| time: %s |", stringutil.MaxFirstCharsWithDots(logTime, trimmedWidth))
		}
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", idRow, versionRow, collectionRow, timeRow)
}

func getRunningStepFooterMainSection(stepRunResult models.StepRunResultsModel) string {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth - 1

	icon := ""
	title := getTrimmedStepName(stepRunResult)
	runTimeStr := TimeToFormattedSeconds(stepRunResult.RunTime, " sec")
	coloringFunc := colorstring.Green

	switch stepRunResult.Status {
	case models.StepRunStatusCodeSuccess:
		icon = "✅"
		coloringFunc = colorstring.Green
		break
	case models.StepRunStatusCodeFailed:
		icon = "🚫"
		coloringFunc = colorstring.Red
		break
	case models.StepRunStatusCodeFailedSkippable:
		icon = "⚠️"
		coloringFunc = colorstring.Yellow
		break
	case models.StepRunStatusCodeSkipped, models.StepRunStatusCodeSkippedWithRunIf:
		icon = "➡"
		coloringFunc = colorstring.Blue
		break
	default:
		log.Error("Unkown result code")
		return ""
	}

	iconBox := fmt.Sprintf(" %s  ", icon)

	titleWhiteSpaceWidth := titleBoxWidth - len(title)
	coloredTitle := title
	if strings.HasPrefix(title, "[Deprecated]") {
		title := strings.TrimPrefix(title, "[Deprecated]")
		coloredTitle = fmt.Sprintf("%s%s", colorstring.Red("[Deprecated]"), coloringFunc(title))
	} else {
		coloredTitle = coloringFunc(title)
	}

	titleBox := fmt.Sprintf(" %s%s", coloredTitle, strings.Repeat(" ", titleWhiteSpaceWidth))

	timeWhiteSpaceWidth := timeBoxWidth - len(runTimeStr) - 1
	timeBox := fmt.Sprintf(" %s%s", runTimeStr, strings.Repeat(" ", timeWhiteSpaceWidth))

	return fmt.Sprintf("|%s|%s|%s|", iconBox, titleBox, timeBox)
}

func getDeprecateNotesRows(notes string) string {
	colorDeprecateNote := func(line string) string {
		if strings.HasPrefix(line, "Removal notes:") {
			line = strings.TrimPrefix(line, "Removal notes:")
			line = fmt.Sprintf("%s%s", colorstring.Red("Removal notes:"), line)
		}
		return line
	}

	boxContentWidth := stepRunSummaryBoxWidthInChars - 4

	notesWithoutNewLine := strings.Replace(notes, "\n", " ", -1)
	words := strings.Split(notesWithoutNewLine, " ")
	if len(words) == 0 {
		return ""
	}

	formattedNote := ""
	line := ""

	for i, word := range words {
		isLastLine := (i == len(words)-1)

		expectedLine := ""
		if line == "" {
			expectedLine = word
		} else {
			expectedLine = line + " " + word
		}

		if utf8.RuneCountInString(expectedLine) > boxContentWidth {
			// expected line would be to long, so print the previous line, and start a new with the last word.
			noteRow := fmt.Sprintf("| %s |", line)
			charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
			if charDiff <= 0 {
				// shorter than desired - fill with space
				line = colorDeprecateNote(line)
				noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
			} else if charDiff > 0 {
				// longer than desired - should not
				log.Errorln("Should not be longer then expected")
			}

			if formattedNote == "" {
				formattedNote = noteRow
			} else {
				formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
			}

			line = word

			if isLastLine {
				noteRow := fmt.Sprintf("| %s |", line)
				charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
				if charDiff < 0 {
					// shorter than desired - fill with space
					line = colorDeprecateNote(line)
					noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
				} else if charDiff > 0 {
					// longer than desired - should not
					log.Errorln("Should not be longer then expected")
				}

				if formattedNote == "" {
					formattedNote = noteRow
				} else {
					formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
				}
			}
		} else {
			// expected line is not to long, just keep growing the line
			line = expectedLine

			if isLastLine {
				noteRow := fmt.Sprintf("| %s |", line)
				charDiff := len(noteRow) - stepRunSummaryBoxWidthInChars
				if charDiff <= 0 {
					// shorter than desired - fill with space
					line = colorDeprecateNote(line)
					noteRow = fmt.Sprintf("| %s%s |", line, strings.Repeat(" ", -charDiff))
				} else if charDiff > 0 {
					// longer than desired - should not
					log.Errorln("Should not be longer then expected")
				}

				if formattedNote == "" {
					formattedNote = noteRow
				} else {
					formattedNote = fmt.Sprintf("%s\n%s", formattedNote, noteRow)
				}
			}
		}
	}

	return formattedNote
}

func getRunningStepFooterSubSection(stepRunResult models.StepRunResultsModel) string {
	stepInfo := stepRunResult.StepInfo

	removalDate := stepInfo.GlobalInfo.RemovalDate
	deprecateNotes := stepInfo.GlobalInfo.DeprecateNotes
	removalDateRow := ""
	deprecateNotesRow := ""
	if removalDate != "" {
		removalDateValue := removalDate
		removalDateKey := colorstring.Red("Removal date:")

		removalDateRow = fmt.Sprintf("| Removal date: %s |", removalDateValue)
		charDiff := len(removalDateRow) - stepRunSummaryBoxWidthInChars
		removalDateRow = fmt.Sprintf("| %s %s%s |", removalDateKey, removalDateValue, strings.Repeat(" ", -charDiff))

		if deprecateNotes != "" {
			deprecateNotesStr := fmt.Sprintf("Removal notes: %s", deprecateNotes)
			deprecateNotesRow = getDeprecateNotesRows(deprecateNotesStr)
		}
	}

	isUpdateAvailable := isUpdateAvailable(stepRunResult.StepInfo)
	updateRow := ""
	if isUpdateAvailable {
		updateRow = fmt.Sprintf("| Update available: %s -> %s |", stepInfo.Version, stepInfo.Latest)
		charDiff := len(updateRow) - stepRunSummaryBoxWidthInChars
		if charDiff < 0 {
			// shorter than desired - fill with space
			updateRow = fmt.Sprintf("| Update available: %s -> %s%s |", stepInfo.Version, stepInfo.Latest, strings.Repeat(" ", -charDiff))
		} else if charDiff > 0 {
			// longer than desired - trim title
			if charDiff > 6 {
				updateRow = fmt.Sprintf("| Update available!%s |", strings.Repeat(" ", -len("| Update available! |")-stepRunSummaryBoxWidthInChars))
			} else {
				updateRow = fmt.Sprintf("| Update available: -> %s%s |", stepInfo.Latest, strings.Repeat(" ", -len("| Update available: -> %s |")-stepRunSummaryBoxWidthInChars))
			}
		}
	}

	issueRow := ""
	sourceRow := ""
	if stepRunResult.Error != nil {
		issueRow = fmt.Sprintf("| Issue tracker: %s |", stepInfo.SupportURL)
		if stepInfo.SupportURL != "" {
			charDiff := len(issueRow) - stepRunSummaryBoxWidthInChars
			if charDiff < 0 {
				// shorter than desired - fill with space
				issueRow = fmt.Sprintf("| Issue tracker: %s%s |", stepInfo.SupportURL, strings.Repeat(" ", -charDiff))
			} else if charDiff > 0 {
				// longer than desired - trim title
				trimmedWidth := len(stepInfo.SupportURL) - charDiff
				if trimmedWidth < 4 {
					log.Errorf("Support url too long, can't present support url at all! : %s", stepInfo.SupportURL)
				} else {
					issueRow = fmt.Sprintf("| Issue tracker: %s |", stringutil.MaxLastCharsWithDots(stepInfo.SupportURL, trimmedWidth))
				}
			}
		}

		sourceRow = fmt.Sprintf("| Source: %s |", stepInfo.SourceCodeURL)
		if stepInfo.SourceCodeURL != "" {
			charDiff := len(sourceRow) - stepRunSummaryBoxWidthInChars
			if charDiff < 0 {
				// shorter than desired - fill with space
				sourceRow = fmt.Sprintf("| Source: %s%s |", stepInfo.SourceCodeURL, strings.Repeat(" ", -charDiff))
			} else if charDiff > 0 {
				// longer than desired - trim title
				trimmedWidth := len(stepInfo.SourceCodeURL) - charDiff
				if trimmedWidth < 4 {
					log.Errorf("Source url too long, can't present source url at all! : %s", stepInfo.SourceCodeURL)
				} else {
					sourceRow = fmt.Sprintf("| Source: %s |", stringutil.MaxLastCharsWithDots(stepInfo.SourceCodeURL, trimmedWidth))
				}
			}
		}
	}

	content := ""
	if isUpdateAvailable {
		content = fmt.Sprintf("%s", updateRow)
	}
	if stepInfo.SupportURL != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, issueRow)
		} else {
			content = fmt.Sprintf("%s", issueRow)
		}
	}
	if stepInfo.SourceCodeURL != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, sourceRow)
		} else {
			content = fmt.Sprintf("%s", sourceRow)
		}
	}
	if removalDate != "" {
		if content != "" {
			content = fmt.Sprintf("%s\n%s", content, removalDateRow)
		} else {
			content = fmt.Sprintf("%s", removalDateRow)
		}

		if deprecateNotes != "" {
			if content != "" {
				content = fmt.Sprintf("%s\n%s", content, deprecateNotesRow)
			} else {
				content = fmt.Sprintf("%s", deprecateNotesRow)
			}
		}
	}

	return content
}

// PrintRunningStepHeader ...
func PrintRunningStepHeader(stepInfo stepmanModels.StepInfoModel, idx int) {
	sep := fmt.Sprintf("+%s+", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	fmt.Println(sep)
	fmt.Println(getRunningStepHeaderMainSection(stepInfo, idx))
	fmt.Println(sep)
	fmt.Println(getRunningStepHeaderSubSection(stepInfo))
	fmt.Println(sep)
	fmt.Println("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")
}

// PrintRunningStepFooter ..
func PrintRunningStepFooter(stepRunResult models.StepRunResultsModel, isLastStepInWorkflow bool) {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth
	sep := fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	fmt.Println("|" + strings.Repeat(" ", stepRunSummaryBoxWidthInChars-2) + "|")

	fmt.Println(sep)
	fmt.Println(getRunningStepFooterMainSection(stepRunResult))
	fmt.Println(sep)
	if stepRunResult.Error != nil || stepRunResult.StepInfo.GlobalInfo.RemovalDate != "" {
		footerSubSection := getRunningStepFooterSubSection(stepRunResult)
		if footerSubSection != "" {
			fmt.Println(footerSubSection)
			fmt.Println(sep)
		}
	}

	if !isLastStepInWorkflow {
		fmt.Println()
		fmt.Println(strings.Repeat(" ", 42) + "▼")
		fmt.Println()
	}
}

// PrintRunningWorkflow ...
func PrintRunningWorkflow(title string) {
	fmt.Println()
	log.Info(colorstring.Bluef("Running workflow (%s)", title))
	fmt.Println()
}

// PrintSummary ...
func PrintSummary(buildRunResults models.BuildRunResultsModel) {
	iconBoxWidth := len("    ")
	timeBoxWidth := len(" time (s) ")
	titleBoxWidth := stepRunSummaryBoxWidthInChars - 4 - iconBoxWidth - timeBoxWidth

	fmt.Println()
	fmt.Println()
	fmt.Printf("+%s+\n", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))
	whitespaceWidth := (stepRunSummaryBoxWidthInChars - 2 - len("bitrise summary ")) / 2
	fmt.Printf("|%sbitrise summary %s|\n", strings.Repeat(" ", whitespaceWidth), strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	whitespaceWidth = stepRunSummaryBoxWidthInChars - len("|    | title") - len("| time (s) |")
	fmt.Printf("|    | title%s| time (s) |\n", strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))

	orderedResults := buildRunResults.OrderedResults()
	tmpTime := time.Time{}
	for _, stepRunResult := range orderedResults {
		tmpTime = tmpTime.Add(stepRunResult.RunTime)
		fmt.Println(getRunningStepFooterMainSection(stepRunResult))
		fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))
		if stepRunResult.Error != nil || stepRunResult.StepInfo.GlobalInfo.RemovalDate != "" {
			footerSubSection := getRunningStepFooterSubSection(stepRunResult)
			if footerSubSection != "" {
				fmt.Println(footerSubSection)
				fmt.Printf("+%s+%s+%s+\n", strings.Repeat("-", iconBoxWidth), strings.Repeat("-", titleBoxWidth), strings.Repeat("-", timeBoxWidth))
			}
		}
	}
	runtime := tmpTime.Sub(time.Time{})

	runtimeStr := TimeToFormattedSeconds(runtime, " sec")
	whitespaceWidth = stepRunSummaryBoxWidthInChars - len(fmt.Sprintf("| Total runtime: %s|", runtimeStr))
	fmt.Printf("| Total runtime: %s%s|\n", runtimeStr, strings.Repeat(" ", whitespaceWidth))
	fmt.Printf("+%s+\n", strings.Repeat("-", stepRunSummaryBoxWidthInChars-2))

	fmt.Println()
}
