package web

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	loginMaxFailures    = 5                // неверных попыток с одного IP до блокировки
	loginFailureWindow  = 10 * time.Minute // за какое время считаем попытки
	loginBlockDuration  = 15 * time.Minute // сколько IP заблокирован
	loginNotifyInterval = 10 * time.Minute // как часто сообщаем админам о неверном пароле с одного IP
	loginCleanInterval  = time.Hour        // как часто чистим память от старых IP
)

type loginAttempts struct {
	failures     []time.Time
	blockedUntil time.Time
	lastNotify   time.Time
}

// loginGuard считает неверные пароли по IP, блокирует перебор и решает, когда сообщать админам
type loginGuard struct {
	mu        sync.Mutex
	byIP      map[string]*loginAttempts
	lastClean time.Time
	loc       *time.Location
}

func newLoginGuard(loc *time.Location) *loginGuard {
	return &loginGuard{
		byIP: make(map[string]*loginAttempts),
		loc:  loc,
	}
}

// blockedFor возвращает, сколько ещё IP заблокирован, 0 если не заблокирован
func (g *loginGuard) blockedFor(ip string, now time.Time) time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()

	a := g.byIP[ip]
	if a == nil || !now.Before(a.blockedUntil) {
		return 0
	}

	return a.blockedUntil.Sub(now)
}

// fail учитывает неверный пароль, notify - текст для админов или пустая строка, если сообщать не надо
func (g *loginGuard) fail(ip string, user string, now time.Time) (notify string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.clean(now)

	a := g.byIP[ip]
	if a == nil {
		a = &loginAttempts{}
		g.byIP[ip] = a
	}

	// оставляем попытки только за последние loginFailureWindow
	recent := a.failures[:0]
	for _, t := range a.failures {
		if now.Sub(t) < loginFailureWindow {
			recent = append(recent, t)
		}
	}
	a.failures = append(recent, now)

	when := now.In(g.loc).Format("02.01.2006 15:04")
	user = shortUser(user)

	if len(a.failures) >= loginMaxFailures {
		a.blockedUntil = now.Add(loginBlockDuration)
		a.failures = nil
		a.lastNotify = now
		return fmt.Sprintf(
			"Веб-модуль: IP %s заблокирован на %d минут после %d неверных паролей за %d минут.\nИмя: %s\nВремя: %s",
			ip, int(loginBlockDuration.Minutes()), loginMaxFailures, int(loginFailureWindow.Minutes()), user, when,
		)
	}

	if a.lastNotify.IsZero() || now.Sub(a.lastNotify) >= loginNotifyInterval {
		a.lastNotify = now
		return fmt.Sprintf(
			"Веб-модуль: неверный пароль.\nIP: %s\nИмя: %s\nВремя: %s\nНеверных попыток с этого IP за %d минут: %d",
			ip, user, when, int(loginFailureWindow.Minutes()), len(a.failures),
		)
	}

	return ""
}

// clean забывает IP, по которым давно ничего не происходило
func (g *loginGuard) clean(now time.Time) {
	if now.Sub(g.lastClean) < loginCleanInterval {
		return
	}
	g.lastClean = now

	for ip, a := range g.byIP {
		idleSince := a.lastNotify
		if n := len(a.failures); n > 0 && a.failures[n-1].After(idleSince) {
			idleSince = a.failures[n-1]
		}
		if now.Sub(idleSince) > loginFailureWindow && !now.Before(a.blockedUntil) {
			delete(g.byIP, ip)
		}
	}
}

// shortUser обрезает введённое имя: оно приходит от кого угодно
func shortUser(user string) string {
	runes := []rune(user)
	if len(runes) > 50 {
		return string(runes[:50]) + "…"
	}
	if len(runes) == 0 {
		return "(пусто)"
	}

	return user
}

// clientIP берёт IP клиента. За обратным прокси (Caddy) это последний адрес в X-Forwarded-For,
// напрямую - адрес соединения. Порт бота наружу не публикуется, подделать заголовок может только прокси.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[len(parts)-1])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
