package web

import (
	"bytes"
	"context"
	"crypto/subtle"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"workout-bot/internal/config"
	"workout-bot/internal/model"
	"workout-bot/internal/service"
)

//go:embed templates/*.html
var templatesFS embed.FS

// pageData - данные для всех страниц
type pageData struct {
	Msg          string
	TimeZone     string
	SendTime     string
	Participants []model.Participant
	Challenges   []model.DocWorkout
	Posts        []model.ChannelPostInfo
}

// PaymentNotifier сообщает человеку в Telegram, что оплата подтверждена
type PaymentNotifier interface {
	NotifyPaymentConfirmed(ctx context.Context, c model.PaymentConfirmation) error
}

type Server struct {
	cfg                   config.WebConfig
	timeZone              string
	userService           *service.UserService
	docWorkoutService     *service.DocWorkoutService
	docTabWorkoutService  *service.DocTabWorkoutService
	docUserWorkoutService *service.DocUserWorkoutService
	channelPostService    *service.ChannelPostService
	settingsService       *service.SettingsService
	notifier              PaymentNotifier
	onSendTimeChanged     func()
	pages                 map[string]*template.Template
}

func NewServer(
	cfg config.WebConfig,
	timeZone string,
	userService *service.UserService,
	docWorkoutService *service.DocWorkoutService,
	docTabWorkoutService *service.DocTabWorkoutService,
	docUserWorkoutService *service.DocUserWorkoutService,
	channelPostService *service.ChannelPostService,
	settingsService *service.SettingsService,
	notifier PaymentNotifier,
	onSendTimeChanged func(),
) (*Server, error) {

	funcs := template.FuncMap{
		"mb": func(b int64) string {
			return fmt.Sprintf("%.1f МБ", float64(b)/(1024*1024))
		},
		"mediaLabel": func(mediaType string) string {
			switch mediaType {
			case "video":
				return "видео"
			case "document":
				return "файл"
			case "photo":
				return "фото"
			case "text":
				return "текст"
			}
			return "другое"
		},
	}

	pages := make(map[string]*template.Template)
	for _, name := range []string{"users.html", "challenges.html", "days.html"} {
		t, err := template.New(name).Funcs(funcs).ParseFS(templatesFS, "templates/layout.html", "templates/"+name)
		if err != nil {
			return nil, err
		}
		pages[name] = t
	}

	return &Server{
		cfg:                   cfg,
		timeZone:              timeZone,
		userService:           userService,
		docWorkoutService:     docWorkoutService,
		docTabWorkoutService:  docTabWorkoutService,
		docUserWorkoutService: docUserWorkoutService,
		channelPostService:    channelPostService,
		settingsService:       settingsService,
		notifier:              notifier,
		onSendTimeChanged:     onSendTimeChanged,
		pages:                 pages,
	}, nil
}

func (s *Server) Run(ctx context.Context) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/users", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /users", s.handleUsers)
	mux.HandleFunc("POST /users/{id}/paid", s.handleUserPaid)
	mux.HandleFunc("GET /challenges", s.handleChallenges)
	mux.HandleFunc("POST /challenges", s.handleChallengeCreate)
	mux.HandleFunc("GET /days", s.handleDays)
	mux.HandleFunc("POST /days", s.handleDaySave)
	mux.HandleFunc("POST /days/send-time", s.handleSendTimeSave)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", s.cfg.Port),
		Handler:           s.basicAuth(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		slog.Info("Stopping web server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Web server shutdown", "Error", err)
		}
		close(done)
	}()

	slog.Info("Web server started", "addr", srv.Addr)
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		<-done
		return
	}
	if err != nil {
		slog.Error("Web server", "Error", err)
	}
}

// basicAuth пускает по одному паролю, имя пользователя любое
func (s *Server) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(pass), []byte(s.cfg.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="workout"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) render(w http.ResponseWriter, page string, data pageData) {
	var buf bytes.Buffer
	if err := s.pages[page].ExecuteTemplate(&buf, "layout", data); err != nil {
		slog.Error("Server.render", "page", page, "Error", err)
		http.Error(w, "Ошибка страницы", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// redirectMsg возвращает на страницу с сообщением
func redirectMsg(w http.ResponseWriter, r *http.Request, path string, msg string) {
	http.Redirect(w, r, path+"?msg="+url.QueryEscape(msg), http.StatusSeeOther)
}
