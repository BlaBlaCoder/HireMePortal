package signupapiv1

import (
	"encoding/json"
	"math"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ledongthuc/pdf"
)

// --- Structs for JSON Response ---

type SkillRating struct {
	Skill    string `json:"skill"`
	Mentions int    `json:"mentions"`
	Rating   string `json:"rating"`
}

type ContactInfo struct {
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// OverallRating holds the final calculated score and a descriptive label.
type OverallRating struct {
	ScorePercentage float64 `json:"scorePercentage"`
	RatingLabel     string  `json:"ratingLabel"`
}

// AnalysisResponse is the top-level structure for the API's JSON response.
type AnalysisResponse struct {
	Overall OverallRating `json:"overallRating"`
	Skills  []SkillRating `json:"skills"`
	Contact ContactInfo   `json:"contactInfo"`
}

// WeightedSkill defines a skill and its importance (weight).
type WeightedSkill struct {
	Keyword string
	Weight  int // A higher number means more important
}

// --- Analysis Logic ---

// getWeightedSkills defines the list of skills and their importance.
// This is the main place to configure the rating logic.
func getWeightedSkills() []WeightedSkill {
	return []WeightedSkill{
		// High Importance (Weight 5) - Core skills and domains
		{Keyword: "product", Weight: 5},
		{Keyword: "api", Weight: 5},
		{Keyword: "payment", Weight: 5},
		{Keyword: "agile", Weight: 5},

		// Medium Importance (Weight 3) - Key technologies and specific experience
		{Keyword: "genai", Weight: 3},
		{Keyword: "aws", Weight: 3},
		{Keyword: "gds", Weight: 3},
		{Keyword: "pss", Weight: 3},
		{Keyword: "iso 20022", Weight: 3},

		// Standard Importance (Weight 1) - Common tools and supporting skills
		{Keyword: "sql", Weight: 1},
		{Keyword: "jira", Weight: 1},
		{Keyword: "llm", Weight: 1},
		{Keyword: "python", Weight: 1},
	}
}

func extractTextFromPDF1(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var textBuilder strings.Builder
	numPages := reader.NumPage()

	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		textBuilder.WriteString(text)
	}
	return textBuilder.String(), nil
}

func getSkillRating(count int) string {
	switch {
	case count >= 5:
		return "Expert / Heavily Utilized ⭐⭐⭐"
	case count >= 3:
		return "Proficient ⭐⭐"
	case count > 0:
		return "Familiar ⭐"
	default:
		return "Not Found"
	}
}

func getOverallRatingLabel(percentage float64) string {
	switch {
	case percentage >= 75:
		return "Excellent Match"
	case percentage >= 60:
		return "Strong Match"
	case percentage >= 40:
		return "Good Match"
	default:
		return "Potential Match"
	}
}

// --- API Handler ---

func AnalyzeResumeHandler(w http.ResponseWriter, r *http.Request) {
	resumePath := "resumes/user_7_5143f376-f529-4b7d-815b-9420e0106b69_20251005.pdf"

	if _, err := os.Stat(resumePath); os.IsNotExist(err) {
		http.Error(w, "Resume PDF file not found on server", http.StatusInternalServerError)
		return
	}

	resumeText, err := extractTextFromPDF1(resumePath)
	if err != nil {
		http.Error(w, "Failed to parse resume PDF", http.StatusInternalServerError)
		return
	}

	resumeTextLower := strings.ToLower(resumeText)

	// --- Perform Analysis ---
	var response AnalysisResponse
	weightedSkills := getWeightedSkills()

	var totalScore float64
	var maxPossibleScore float64
	const mentionCap = 5 // Cap mentions for a skill to prevent keyword stuffing

	for _, skill := range weightedSkills {
		count := strings.Count(resumeTextLower, skill.Keyword)

		// Add individual skill rating to the list
		response.Skills = append(response.Skills, SkillRating{
			Skill:    skill.Keyword,
			Mentions: count,
			Rating:   getSkillRating(count),
		})

		// Calculate weighted score
		cappedCount := math.Min(float64(count), float64(mentionCap))
		totalScore += cappedCount * float64(skill.Weight)
		maxPossibleScore += float64(mentionCap) * float64(skill.Weight)
	}

	// Calculate final percentage and get label
	scorePercentage := 0.0
	if maxPossibleScore > 0 {
		scorePercentage = (totalScore / maxPossibleScore) * 100
	}
	response.Overall.ScorePercentage = math.Round(scorePercentage*100) / 100 // Round to 2 decimal places
	response.Overall.RatingLabel = getOverallRatingLabel(scorePercentage)

	// Extract Contact Info
	emailRegex := regexp.MustCompile(`[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	response.Contact.Email = emailRegex.FindString(resumeTextLower)

	phoneRegex := regexp.MustCompile(`\+91-?\d{10}`)
	response.Contact.Phone = phoneRegex.FindString(resumeText)

	// --- Send JSON Response ---
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
