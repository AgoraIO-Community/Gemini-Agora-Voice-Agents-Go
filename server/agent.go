package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	Agora "github.com/AgoraIO/agora-agents-go/v2"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit"
	"github.com/AgoraIO/agora-agents-go/v2/agentkit/vendors"
	"github.com/AgoraIO/agora-agents-go/v2/option"
)

const adaPrompt = `You are Ada, an agentic developer advocate from Agora. You help developers understand and build with Agora's Conversational AI platform.

Agora is a real-time communications company. The product you represent is the Agora Conversational AI Engine.

If you do not know a specific fact about Agora, say so plainly and suggest checking docs.agora.io. Keep most replies to one or two sentences unless the user explicitly asks for more detail.
`

type agentService struct {
	appID                string
	certificate          string
	greeting             string
	googleAPIKey         string
	sessionClient        *agentkit.AgoraClient

	mu       sync.Mutex
	sessions map[string]sessionStopper
}

type sessionStopper interface {
	Stop(ctx context.Context) error
}

type configData struct {
	AppID       string `json:"app_id"`
	Token       string `json:"token"`
	UID         string `json:"uid"`
	ChannelName string `json:"channel_name"`
	AgentUID    string `json:"agent_uid"`
}

type startAgentResult struct {
	AgentID     string `json:"agent_id"`
	ChannelName string `json:"channel_name"`
	Status      string `json:"status"`
}

func newAgentService() (*agentService, error) {
	appID := strings.TrimSpace(os.Getenv("AGORA_APP_ID"))
	certificate := strings.TrimSpace(os.Getenv("AGORA_APP_CERTIFICATE"))
	if appID == "" || certificate == "" {
		return nil, errors.New("AGORA_APP_ID and AGORA_APP_CERTIFICATE are required")
	}
	googleAPIKey := strings.TrimSpace(os.Getenv("GOOGLE_API_KEY"))
	agoraClient := agentkit.NewAgoraClient(agentkit.AgoraClientOptions{
		Area:           option.AreaUS,
		AppID:          appID,
		AppCertificate: certificate,
	})

	return &agentService{
		appID:                appID,
		certificate:          certificate,
		googleAPIKey:         googleAPIKey,
		greeting: firstNonEmpty(
			strings.TrimSpace(os.Getenv("AGENT_GREETING")),
			"Hi there! I'm Ada, your virtual assistant from Agora. How can I help?",
		),
		sessionClient: agoraClient,
		sessions:      make(map[string]sessionStopper),
	}, nil
}

func (s *agentService) generateConfig(channel string, uid int) (*configData, error) {
	userUID := uid
	if userUID <= 0 {
		userUID = randomInt(1000, 9999999)
	}

	channelName := strings.TrimSpace(channel)
	if channelName == "" {
		channelName = generateChannelName()
	}

	agentUID := randomInt(10000000, 99999999)
	expiry, err := agentkit.ExpiresInHours(1)
	if err != nil {
		return nil, fmt.Errorf("resolve token expiry: %w", err)
	}

	token, err := agentkit.GenerateConvoAIToken(agentkit.GenerateConvoAITokenOptions{
		AppID:          s.appID,
		AppCertificate: s.certificate,
		ChannelName:    channelName,
		UID:            userUID,
		TokenExpire:    expiry,
	})
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &configData{
		AppID:       s.appID,
		Token:       token,
		UID:         strconv.Itoa(userUID),
		ChannelName: channelName,
		AgentUID:    strconv.Itoa(agentUID),
	}, nil
}

func (s *agentService) start(channelName string, agentUID, userUID int) (*startAgentResult, error) {
	channelName = strings.TrimSpace(channelName)
	if channelName == "" {
		return nil, errors.New("channel_name is required and cannot be empty")
	}
	if agentUID <= 0 {
		return nil, errors.New("agent_uid is required and cannot be empty")
	}
	if userUID <= 0 {
		return nil, errors.New("user_uid is required and cannot be empty")
	}
	if s.googleAPIKey == "" {
		return nil, errors.New("GOOGLE_API_KEY is required for GeminiSTT")
	}

	expiresIn, err := agentkit.ExpiresInHours(1)
	if err != nil {
		return nil, fmt.Errorf("resolve session expiry: %w", err)
	}

	enableRTM := true
	enableTools := true
	enableErrorMessage := true
	enableMetrics := true
	dataChannel := agentkit.ParametersDataChannel("rtm")
	enableStringUID := false
	idleTimeout := 30
	speechThreshold := 0.5
	endOfSpeechMode := agentkit.EndOfSpeechMode("vad")
	interruptDurationMs := 160
	prefixPaddingMs := 300
	silenceDurationMs := 480

	agent := agentkit.NewAgent(
		s.sessionClient,
		agentkit.WithInstructions(adaPrompt),
		agentkit.WithGreeting(s.greeting),
		agentkit.WithFailureMessage("Please wait a moment."),
		agentkit.WithTurnDetectionConfig(&agentkit.TurnDetectionConfig{
			Language: Agora.AsrLanguageEnUs.Ptr(),
			Config: &agentkit.TurnDetectionNestedConfig{
				SpeechThreshold: &speechThreshold,
				StartOfSpeech: &agentkit.StartOfSpeechConfig{
					Mode: agentkit.StartOfSpeechMode("vad"),
					VadConfig: &agentkit.StartOfSpeechVadConfig{
						InterruptDurationMs: &interruptDurationMs,
						PrefixPaddingMs:     &prefixPaddingMs,
					},
				},
				EndOfSpeech: &agentkit.EndOfSpeechConfig{
					Mode: &endOfSpeechMode,
					VadConfig: &agentkit.EndOfSpeechVadConfig{
						SilenceDurationMs: &silenceDurationMs,
					},
				},
			},
		}),
		agentkit.WithAdvancedFeatures(&agentkit.AdvancedFeatures{
			EnableRtm:   &enableRTM,
			EnableTools: &enableTools,
		}),
		agentkit.WithParameters(&agentkit.SessionParams{
			DataChannel:        &dataChannel,
			EnableErrorMessage: &enableErrorMessage,
			EnableMetrics:      &enableMetrics,
		}),
		agentkit.WithAudioScenario(agentkit.ParametersAudioScenario("chorus")),
	).
	WithLlm(vendors.NewGemini(vendors.GeminiOptions{
		APIKey:          s.googleAPIKey,
		Model:           "gemini-3.6-flash",
		GreetingMessage: s.greeting,
		FailureMessage:  "Please wait a moment.",
		MaxHistory:      intPtr(15),
	})).
		WithStt(vendors.NewGeminiSTT(vendors.GeminiSTTOptions{
			APIKey:           s.googleAPIKey,
			LanguageCodes:    []string{"en-US"},
			CustomVocabulary: []string{"Agora", "Gemini"},
			WordTimestamp:    Agora.Bool(false),
		})).
	WithTts(vendors.NewMiniMaxTTS(vendors.MiniMaxTTSOptions{
		Model:   "speech_2_6_turbo",
		VoiceID: "English_captivating_female1",
	}))

	session := agent.CreateSession(agentkit.CreateSessionOptions{
		Channel:         channelName,
		AgentUID:        strconv.Itoa(agentUID),
		RemoteUIDs:      []string{strconv.Itoa(userUID)},
		EnableStringUID: &enableStringUID,
		IdleTimeout:     &idleTimeout,
		ExpiresIn:       expiresIn,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	agentID, err := session.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("start agent: %w", err)
	}

	s.mu.Lock()
	s.sessions[agentID] = session
	s.mu.Unlock()

	return &startAgentResult{
		AgentID:     agentID,
		ChannelName: channelName,
		Status:      "started",
	}, nil
}

func (s *agentService) stop(agentID string) error {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return errors.New("agent_id is required and cannot be empty")
	}

	s.mu.Lock()
	session, ok := s.sessions[agentID]
	if ok {
		delete(s.sessions, agentID)
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := session.Stop(ctx); err != nil {
		return fmt.Errorf("stop agent session: %w", err)
	}

	return nil
}

func generateChannelName() string {
	return fmt.Sprintf("ai-conversation-%d-%d", time.Now().Unix(), randomInt(1000, 9999))
}

func randomInt(minInclusive, maxInclusive int) int {
	if maxInclusive <= minInclusive {
		return minInclusive
	}
	return minInclusive + rand.Intn(maxInclusive-minInclusive+1)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func intPtr(value int) *int {
	return &value
}

func float64Ptr(value float64) *float64 {
	return &value
}

func sampleRatePtr(value vendors.SampleRate) *vendors.SampleRate {
	return &value
}
