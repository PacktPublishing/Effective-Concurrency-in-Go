package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

type DashboardService struct {
	Users    UserSvc
	Accounts AccountSvc
}

type UserSvc struct{}

func (UserSvc) GetStats(context.Context, string) (UserStats, error) {
	return UserStats{}, nil
}

func (UserSvc) GetTransactions(context.Context, string) <-chan Transaction {
	return make(chan Transaction)
}

type AccountSvc struct{}

func (AccountSvc) GetStats(context.Context, string) chan AccountStats {
	return make(chan AccountStats)
}

func (AccountSvc) GetTransactions(context.Context, string) <-chan Transaction {
	return make(chan Transaction)
}

type Transaction struct{}

func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

type DashboardParams struct{}

type UserStats struct{}
type AccountStats struct{}

type DashboardData struct {
	UserData         UserStats
	AccountData      AccountStats
	LastTransactions []Transaction
}

func (svc *DashboardService) GetDashboardData(ctx context.Context, userID string) DashboardData {
	result := DashboardData{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		result.UserData, err = svc.Users.GetStats(ctx, userID)
		if err != nil {
			log.Println(err)
		}
	}()
	acctCh := make(chan AccountStats)
	go func() {
		defer close(acctCh)
		newCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		resultCh := svc.Accounts.GetStats(newCtx, userID)
		select {
		case data := <-resultCh:
			acctCh <- data
		case <-newCtx.Done():
		}
	}()

	transactionWg := sync.WaitGroup{}
	transactionWg.Add(1)
	transactionCh := make(chan Transaction)
	go func() {
		defer transactionWg.Done()
		for t := range svc.Users.GetTransactions(ctx, userID) {
			transactionCh <- t
		}
	}()
	transactionWg.Add(1)
	go func() {
		defer transactionWg.Done()
		for t := range svc.Accounts.GetTransactions(ctx, userID) {
			transactionCh <- t
		}
	}()
	go func() {
		transactionWg.Wait()
		close(transactionCh)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for record := range transactionCh {
			result.LastTransactions = append(result.LastTransactions, record)
		}
	}()

	wg.Wait()
	result.AccountData = <-acctCh

	return result
}

func (svc *DashboardService) SetDashboardConfig(ctx context.Context, userID string, params DashboardParams) {
}

func (svc *DashboardService) DashboardHandler(w http.ResponseWriter, req *http.Request) {
	userId := GetUserID(req.Context())
	switch req.Method {
	case http.MethodGet:
		dashboard := svc.GetDashboardData(req.Context(), userId)
		json.NewEncoder(w).Encode(dashboard)
	case http.MethodPost:
		var params DashboardParams
		if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		svc.SetDashboardConfig(req.Context(), userId, params)
	default:
		http.Error(w, "Unhandled request type", http.StatusMethodNotAllowed)
	}
}

func Limit(maxSize int64, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.Body = http.MaxBytesReader(w, req.Body, maxSize)
		next(w, req)
	})
}

type userIDKeyType int

const userIDKey userIDKeyType = iota

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}

func Authenticate(authenticator func(*http.Request) (string, error), next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		userId, err := authenticator(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next(w, req.WithContext(WithUserID(req.Context(), userId)))
	})
}

func ConcurrencyLimiter(sem chan struct{}, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		sem <- struct{}{}
		defer func() { <-sem }()
		next(w, req)
	})
}

func main() {
	mux := http.NewServeMux()
	svc := NewDashboardService()
	mux.HandleFunc("/dashboard/", ConcurrencyLimiter(make(chan struct{}, 20), svc.DashboardHandler))
	http.ListenAndServe("localhost:10001", mux)
}
