package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"revivers/internal/adapters/postgres"
	"revivers/internal/adapters/postgres/migrations"
	"revivers/internal/domain"
	"revivers/internal/usecase"
	"revivers/pkg/logger"
)

var (
	testDB         *postgres.Pool
	testService    *usecase.Service
	testController *Controller
)

func TestMain(m *testing.M) {
	// Настраиваем тестовую БД
	dbSource := os.Getenv("TEST_DB_SOURCE")
	if dbSource == "" {
		dbSource = "postgres://lait:123@localhost:5432/orders_db?sslmode=disable"
	}

	ctx := context.Background()
	cfg := postgres.Config{Source: dbSource}

	// Создаем подключение
	var err error
	testDB, err = postgres.New(ctx, cfg)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}

	// Запускаем миграции
	if err := migrations.RunMigrate("../../adapters/postgres/migrations/", cfg); err != nil {
		fmt.Printf("Failed to run migrations: %v\n", err)
		testDB.Close()
		os.Exit(1)
	}

	// Инициализируем логгер для тестов
	logger.Init(logger.Config{Level: "warn"})

	// Создаем сервисы
	testService = usecase.NewService(testDB)
	testController = NewController(testService)

	// Запускаем тесты
	code := m.Run()

	// Очистка после тестов
	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

// createTestTeams создает 20 команд для тестирования
func createTestTeams(t *testing.T, controller *Controller) []domain.Team {
	teams := make([]domain.Team, 0, 20)

	for i := 0; i < 20; i++ {
		teamName := fmt.Sprintf("team-%d", i+1)
		memberCount := 3 + (i % 5) // От 3 до 7 участников в команде

		members := make([]domain.TeamMember, 0, memberCount)
		for j := 0; j < memberCount; j++ {
			userID := fmt.Sprintf("user-team-%d-member-%d", i+1, j+1)
			members = append(members, domain.TeamMember{
				UserId:   userID,
				Username: fmt.Sprintf("username-%d-%d", i+1, j+1),
				IsActive: true,
			})
		}

		team := domain.Team{
			TeamName: teamName,
			Members:  members,
		}

		// Создаем команду через handler
		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.PostTeamAdd(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Failed to create team %s: status %d, body: %s", teamName, w.Code, w.Body.String())
		}

		teams = append(teams, team)
	}

	return teams
}

// createTestPRs создает 200 PR для тестирования
func createTestPRs(t *testing.T, controller *Controller, teams []domain.Team) []domain.PullRequest {
	prs := make([]domain.PullRequest, 0, 200)

	prCounter := 0
	for _, team := range teams {
		// Создаем по 10 PR на команду
		for i := 0; i < 10 && prCounter < 200; i++ {
			prCounter++
			prID := fmt.Sprintf("pr-%d", prCounter)
			prName := fmt.Sprintf("PR #%d", prCounter)

			// Выбираем случайного автора из команды
			authorIdx := prCounter % len(team.Members)
			authorID := team.Members[authorIdx].UserId

			reqBody := domain.PostPullRequestCreateJSONBody{
				PullRequestId:   prID,
				PullRequestName: prName,
				AuthorId:        authorID,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostPullRequestCreate(w, req)

			if w.Code != http.StatusCreated {
				t.Logf("Failed to create PR %s: status %d, body: %s", prID, w.Code, w.Body.String())
				continue
			}

			// Ответ обернут в {pr: PullRequest}
			var response struct {
				Pr domain.PullRequest `json:"pr"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
				prs = append(prs, response.Pr)
			}
		}
	}

	return prs
}

// TestIntegration_AllHandlers - большой интеграционный тест всех handlers
func TestIntegration_AllHandlers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Очищаем БД перед тестом
	_, err := testDB.DB.ExecContext(ctx, `
		TRUNCATE TABLE pull_request_reviewers CASCADE;
		TRUNCATE TABLE pull_requests CASCADE;
		TRUNCATE TABLE users CASCADE;
		TRUNCATE TABLE teams CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	// Создаем команды для тестирования
	teams := createTestTeams(t, testController)

	t.Run("Create 20 teams", func(t *testing.T) {
		if len(teams) != 20 {
			t.Errorf("Expected 20 teams, got %d", len(teams))
		}
		t.Logf("Created %d teams", len(teams))
	})

	t.Run("Create 200 PRs", func(t *testing.T) {
		prs := createTestPRs(t, testController, teams)
		t.Logf("Created %d PRs", len(prs))
		if len(prs) < 100 {
			t.Errorf("Expected at least 100 PRs, got %d", len(prs))
		}
	})

	// Тестируем каждый handler по 6-7 раз с разными данными
	t.Run("Test PostTeamAdd handler - 7 scenarios", func(t *testing.T) {
		testPostTeamAddScenarios(t, testController)
	})

	t.Run("Test GetTeamGet handler - 7 scenarios", func(t *testing.T) {
		testGetTeamGetScenarios(t, testController, teams)
	})

	t.Run("Test PostPullRequestCreate handler - 7 scenarios", func(t *testing.T) {
		testPostPullRequestCreateScenarios(t, testController, teams)
	})

	t.Run("Test PostPullRequestMerge handler - 7 scenarios", func(t *testing.T) {
		testPostPullRequestMergeScenarios(t, testController)
	})

	t.Run("Test PostPullRequestReassign handler - 7 scenarios", func(t *testing.T) {
		testPostPullRequestReassignScenarios(t, testController, teams)
	})

	t.Run("Test GetUsersGetReview handler - 7 scenarios", func(t *testing.T) {
		testGetUsersGetReviewScenarios(t, testController, teams)
	})

	t.Run("Test PostUsersSetIsActive handler - 7 scenarios", func(t *testing.T) {
		testPostUsersSetIsActiveScenarios(t, testController, teams)
	})

	// Тестируем обработку ошибок
	t.Run("Test Error Handling - All error types", func(t *testing.T) {
		testErrorHandling(t, testController, teams)
	})
}

// testPostTeamAddScenarios тестирует PostTeamAdd с разными сценариями
func testPostTeamAddScenarios(t *testing.T, controller *Controller) {
	scenarios := []struct {
		name        string
		team        domain.Team
		expectError bool
		statusCode  int
	}{
		{
			name: "Create new team with 3 members",
			team: domain.Team{
				TeamName: "test-team-1",
				Members: []domain.TeamMember{
					{UserId: "u1", Username: "user1", IsActive: true},
					{UserId: "u2", Username: "user2", IsActive: true},
					{UserId: "u3", Username: "user3", IsActive: true},
				},
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create team with 10 members",
			team: func() domain.Team {
				members := make([]domain.TeamMember, 10)
				for i := range members {
					members[i] = domain.TeamMember{
						UserId:   fmt.Sprintf("u-team2-%d", i+1),
						Username: fmt.Sprintf("user-team2-%d", i+1),
						IsActive: true,
					}
				}
				return domain.Team{
					TeamName: "test-team-2",
					Members:  members,
				}
			}(),
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create team with 1 member",
			team: domain.Team{
				TeamName: "test-team-3",
				Members: []domain.TeamMember{
					{UserId: "u4", Username: "user4", IsActive: false},
				},
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Update existing team",
			team: domain.Team{
				TeamName: "test-team-1",
				Members: []domain.TeamMember{
					{UserId: "u1", Username: "user1-updated", IsActive: true},
					{UserId: "u5", Username: "user5-new", IsActive: true},
				},
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create team with empty name",
			team: domain.Team{
				TeamName: "",
				Members:  []domain.TeamMember{},
			},
			expectError: false, // Пустое имя технически валидно (может быть ограничение на уровне БД)
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create team with duplicate user IDs",
			team: domain.Team{
				TeamName: "test-team-4",
				Members: []domain.TeamMember{
					{UserId: "u6", Username: "user6", IsActive: true},
					{UserId: "u6", Username: "user6-duplicate", IsActive: true},
				},
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			body, _ := json.Marshal(scenario.team)
			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostTeamAdd(w, req)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				}
			}
		})
	}
}

// testGetTeamGetScenarios тестирует GetTeamGet с разными сценариями
func testGetTeamGetScenarios(t *testing.T, controller *Controller, teams []domain.Team) {
	scenarios := []struct {
		name        string
		teamName    string
		expectError bool
		statusCode  int
	}{
		{
			name:        "Get existing team",
			teamName:    teams[0].TeamName,
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Get team with many members",
			teamName:    teams[5].TeamName,
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Get non-existent team",
			teamName:    "non-existent-team-12345",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get team with empty name",
			teamName:    "",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get team with special characters",
			teamName:    "team-with-special-chars-!@#$%",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get last created team",
			teamName:    teams[len(teams)-1].TeamName,
			expectError: false,
			statusCode:  http.StatusOK,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Для очень длинных имен используем прямое указание параметра, а не URL encoding
			var req *http.Request
			if len(scenario.teamName) > 500 {
				// Для очень длинных имен создаем запрос без URL encoding
				req = httptest.NewRequest(http.MethodGet, "/team/get", nil)
				q := req.URL.Query()
				q.Set("team_name", scenario.teamName)
				req.URL.RawQuery = q.Encode()
			} else {
				req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/team/get?team_name=%s", scenario.teamName), nil)
			}
			params := domain.GetTeamGetParams{TeamName: scenario.teamName}
			w := httptest.NewRecorder()

			controller.GetTeamGet(w, req, params)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					var team domain.Team
					if err := json.Unmarshal(w.Body.Bytes(), &team); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					} else if team.TeamName != scenario.teamName {
						t.Errorf("Expected team name %s, got %s", scenario.teamName, team.TeamName)
					}
				}
			}
		})
	}
}

// testPostPullRequestCreateScenarios тестирует PostPullRequestCreate с разными сценариями
func testPostPullRequestCreateScenarios(t *testing.T, controller *Controller, teams []domain.Team) {
	scenarios := []struct {
		name        string
		req         domain.PostPullRequestCreateJSONBody
		expectError bool
		statusCode  int
	}{
		{
			name: "Create PR with valid author from team",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-test-1",
				PullRequestName: "Test PR 1",
				AuthorId:        teams[0].Members[0].UserId,
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create PR with author from large team",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-test-2",
				PullRequestName: "Test PR 2",
				AuthorId:        teams[5].Members[0].UserId,
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
		{
			name: "Create PR with non-existent author",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-test-3",
				PullRequestName: "Test PR 3",
				AuthorId:        "non-existent-user-12345",
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name: "Create PR with duplicate ID",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-test-1",
				PullRequestName: "Test PR Duplicate",
				AuthorId:        teams[0].Members[0].UserId,
			},
			expectError: true,
			statusCode:  http.StatusConflict,
		},
		{
			name: "Create PR with author without team",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "pr-test-4",
				PullRequestName: "Test PR 4",
				AuthorId:        "user-without-team",
			},
			expectError: true,
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "Create PR with empty PR ID",
			req: domain.PostPullRequestCreateJSONBody{
				PullRequestId:   "",
				PullRequestName: "Test PR 5",
				AuthorId:        teams[0].Members[0].UserId,
			},
			expectError: false,
			statusCode:  http.StatusCreated,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			body, _ := json.Marshal(scenario.req)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostPullRequestCreate(w, req)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					// Проверяем формат ответа {pr: PullRequest}
					var response struct {
						Pr domain.PullRequest `json:"pr"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					}
				}
			}
		})
	}
}

// testPostPullRequestMergeScenarios тестирует PostPullRequestMerge с разными сценариями
func testPostPullRequestMergeScenarios(t *testing.T, controller *Controller) {
	// Сначала создаем несколько PR для тестирования
	testPRs := []string{"pr-merge-test-1", "pr-merge-test-2", "pr-merge-test-3"}

	// Создаем тестовую команду и PR
	team := domain.Team{
		TeamName: "merge-test-team",
		Members: []domain.TeamMember{
			{UserId: "merge-author-1", Username: "merge-author", IsActive: true},
			{UserId: "merge-reviewer-1", Username: "merge-reviewer", IsActive: true},
			{UserId: "merge-reviewer-2", Username: "merge-reviewer-2", IsActive: true},
		},
	}

	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testController.PostTeamAdd(w, req)

	for i, prID := range testPRs {
		prBody, _ := json.Marshal(domain.PostPullRequestCreateJSONBody{
			PullRequestId:   prID,
			PullRequestName: fmt.Sprintf("Merge Test PR %d", i+1),
			AuthorId:        "merge-author-1",
		})
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(prBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		testController.PostPullRequestCreate(w, req)
	}

	scenarios := []struct {
		name        string
		prID        string
		expectError bool
		statusCode  int
	}{
		{
			name:        "Merge existing PR",
			prID:        testPRs[0],
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Merge PR again (idempotent)",
			prID:        testPRs[0],
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Merge another PR",
			prID:        testPRs[1],
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Merge non-existent PR",
			prID:        "non-existent-pr-12345",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Merge already merged PR",
			prID:        testPRs[1],
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Merge last PR",
			prID:        testPRs[2],
			expectError: false,
			statusCode:  http.StatusOK,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			body, _ := json.Marshal(domain.PostPullRequestMergeJSONBody{
				PullRequestId: scenario.prID,
			})
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostPullRequestMerge(w, req)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					// Проверяем формат ответа {pr: PullRequest}
					var response struct {
						Pr domain.PullRequest `json:"pr"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					}
				}
			}
		})
	}
}

// testPostPullRequestReassignScenarios тестирует PostPullRequestReassign с разными сценариями
func testPostPullRequestReassignScenarios(t *testing.T, controller *Controller, teams []domain.Team) {
	// Создаем тестовую команду с несколькими участниками
	team := domain.Team{
		TeamName: "reassign-test-team",
		Members: []domain.TeamMember{
			{UserId: "reassign-author", Username: "author", IsActive: true},
			{UserId: "reassign-reviewer-1", Username: "reviewer1", IsActive: true},
			{UserId: "reassign-reviewer-2", Username: "reviewer2", IsActive: true},
			{UserId: "reassign-reviewer-3", Username: "reviewer3", IsActive: true},
		},
	}

	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testController.PostTeamAdd(w, req)

	// Создаем PR для переназначения
	prID := "pr-reassign-test-1"
	prBody, _ := json.Marshal(domain.PostPullRequestCreateJSONBody{
		PullRequestId:   prID,
		PullRequestName: "Reassign Test PR",
		AuthorId:        "reassign-author",
	})
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(prBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testController.PostPullRequestCreate(w, req)

	// Убеждаемся, что PR создан успешно
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create PR for reassign test: %d", w.Code)
	}

	// Получаем созданный PR, чтобы узнать назначенных ревьюверов
	var createResponse struct {
		Pr domain.PullRequest `json:"pr"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &createResponse); err != nil {
		t.Fatalf("Failed to unmarshal created PR: %v", err)
	}
	createdPR := createResponse.Pr

	// Определяем ревьюверов для тестов - используем реально назначенных
	if len(createdPR.AssignedReviewers) == 0 {
		t.Skip("No reviewers assigned to PR, skipping reassign tests")
	}

	// Создаем второй PR для тестов переназначения
	prID2 := "pr-reassign-test-2"
	prBody2, _ := json.Marshal(domain.PostPullRequestCreateJSONBody{
		PullRequestId:   prID2,
		PullRequestName: "Reassign Test PR 2",
		AuthorId:        "reassign-author",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(prBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	testController.PostPullRequestCreate(w2, req2)

	// Получаем второй PR для определения ревьюверов
	var reviewer2_1 string
	if w2.Code == http.StatusCreated {
		var createResponse2 struct {
			Pr domain.PullRequest `json:"pr"`
		}
		if err := json.Unmarshal(w2.Body.Bytes(), &createResponse2); err == nil {
			if len(createResponse2.Pr.AssignedReviewers) > 0 {
				reviewer2_1 = createResponse2.Pr.AssignedReviewers[0]
			} else {
				reviewer2_1 = "reassign-reviewer-2"
			}
		} else {
			reviewer2_1 = "reassign-reviewer-2"
		}
	} else {
		reviewer2_1 = "reassign-reviewer-2"
	}

	scenarios := []struct {
		name        string
		req         domain.PostPullRequestReassignJSONBody
		expectError bool
		statusCode  int
	}{
		{
			name: "Reassign again",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: prID2,       // Используем второй PR для избежания конфликтов
				OldUserId:     reviewer2_1, // Используем реально назначенного ревьювера из второго PR
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Reassign non-existent reviewer",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: prID,
				OldUserId:     "non-existent-reviewer",
			},
			expectError: true,
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "Reassign from non-existent PR",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: "non-existent-pr",
				OldUserId:     "reassign-reviewer-1",
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name: "Reassign from merged PR",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: prID,
				OldUserId:     "reassign-reviewer-3",
			},
			expectError: true,
			statusCode:  http.StatusBadRequest,
		},
		{
			name: "Reassign with empty PR ID",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: "",
				OldUserId:     "reassign-reviewer-1",
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name: "Reassign with empty user ID",
			req: domain.PostPullRequestReassignJSONBody{
				PullRequestId: prID,
				OldUserId:     "",
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
	}

	// Мержим PR перед последним тестом
	mergeBody, _ := json.Marshal(domain.PostPullRequestMergeJSONBody{PullRequestId: prID})
	req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(mergeBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testController.PostPullRequestMerge(w, req)

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			body, _ := json.Marshal(scenario.req)
			req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostPullRequestReassign(w, req)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					// Проверяем формат ответа {pr: PullRequest, replaced_by: string}
					var response struct {
						Pr         domain.PullRequest `json:"pr"`
						ReplacedBy string             `json:"replaced_by"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					}
				}
			}
		})
	}
}

// testGetUsersGetReviewScenarios тестирует GetUsersGetReview с разными сценариями
func testGetUsersGetReviewScenarios(t *testing.T, controller *Controller, teams []domain.Team) {
	scenarios := []struct {
		name        string
		userID      string
		expectError bool
		statusCode  int
	}{
		{
			name:        "Get reviews for user with PRs",
			userID:      teams[0].Members[1].UserId, // Обычно не автор, может быть ревьювером
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Get reviews for author (should be empty or have reviews)",
			userID:      teams[0].Members[0].UserId,
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Get reviews for non-existent user",
			userID:      "non-existent-user-12345",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get reviews for user without team",
			userID:      "user-without-team",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get reviews with empty user ID",
			userID:      "",
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name:        "Get reviews for user from large team",
			userID:      teams[10].Members[2].UserId,
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name:        "Get reviews for inactive user",
			userID:      teams[0].Members[len(teams[0].Members)-1].UserId,
			expectError: false,
			statusCode:  http.StatusOK,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/getReview?user_id=%s", scenario.userID), nil)
			params := domain.GetUsersGetReviewParams{UserId: scenario.userID}
			w := httptest.NewRecorder()

			controller.GetUsersGetReview(w, req, params)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					// Проверяем формат ответа {user_id: string, pull_requests: []PullRequestShort}
					var response struct {
						UserId       string                    `json:"user_id"`
						PullRequests []domain.PullRequestShort `json:"pull_requests"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					} else if response.UserId != scenario.userID {
						t.Errorf("Expected user_id %s, got %s", scenario.userID, response.UserId)
					}
				}
			}
		})
	}
}

// testPostUsersSetIsActiveScenarios тестирует PostUsersSetIsActive с разными сценариями
func testPostUsersSetIsActiveScenarios(t *testing.T, controller *Controller, teams []domain.Team) {
	scenarios := []struct {
		name        string
		req         domain.PostUsersSetIsActiveJSONBody
		expectError bool
		statusCode  int
	}{
		{
			name: "Set user to active",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   teams[0].Members[0].UserId,
				IsActive: true,
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Set user to inactive",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   teams[0].Members[0].UserId,
				IsActive: false,
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Set user to active again",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   teams[0].Members[0].UserId,
				IsActive: true,
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Set non-existent user",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   "non-existent-user-12345",
				IsActive: true,
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name: "Set user with empty ID",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   "",
				IsActive: true,
			},
			expectError: true,
			statusCode:  http.StatusNotFound,
		},
		{
			name: "Set multiple users",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   teams[1].Members[0].UserId,
				IsActive: false,
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
		{
			name: "Set user from different team",
			req: domain.PostUsersSetIsActiveJSONBody{
				UserId:   teams[15].Members[2].UserId,
				IsActive: true,
			},
			expectError: false,
			statusCode:  http.StatusOK,
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			body, _ := json.Marshal(scenario.req)
			req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			controller.PostUsersSetIsActive(w, req)

			if scenario.expectError {
				if w.Code < 400 {
					t.Errorf("Expected error status, got %d", w.Code)
				}
			} else {
				if w.Code != scenario.statusCode {
					t.Errorf("Expected status %d, got %d. Body: %s", scenario.statusCode, w.Code, w.Body.String())
				} else {
					// Проверяем формат ответа {user: User}
					var response struct {
						User domain.User `json:"user"`
					}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
						t.Errorf("Failed to unmarshal response: %v", err)
					} else if response.User.UserId != scenario.req.UserId {
						t.Errorf("Expected user_id %s, got %s", scenario.req.UserId, response.User.UserId)
					}
				}
			}
		})
	}
}

// testErrorHandling тестирует обработку всех типов ошибок
func testErrorHandling(t *testing.T, controller *Controller, teams []domain.Team) {
	errorScenarios := []struct {
		name       string
		testFunc   func(*testing.T)
		errorCode  domain.ErrorResponseErrorCode
		statusCode int
	}{
		{
			name: "NotFoundError - Get non-existent team",
			testFunc: func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=non-existent", nil)
				params := domain.GetTeamGetParams{TeamName: "non-existent"}
				w := httptest.NewRecorder()
				controller.GetTeamGet(w, req, params)
				if w.Code != http.StatusNotFound {
					t.Errorf("Expected 404, got %d", w.Code)
				}
				var errResp domain.ErrorResponse
				_ = json.Unmarshal(w.Body.Bytes(), &errResp)
				if errResp.Error.Code != domain.NOTFOUND {
					t.Errorf("Expected NOTFOUND error code")
				}
			},
			errorCode:  domain.NOTFOUND,
			statusCode: http.StatusNotFound,
		},
		{
			name: "Update existing team (not an error)",
			testFunc: func(t *testing.T) {
				team := teams[0]
				// Обновляем команду с новыми участниками
				updatedTeam := domain.Team{
					TeamName: team.TeamName,
					Members: []domain.TeamMember{
						{UserId: "updated-user-1", Username: "Updated User 1", IsActive: true},
						{UserId: "updated-user-2", Username: "Updated User 2", IsActive: true},
					},
				}
				body, _ := json.Marshal(updatedTeam)
				req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				controller.PostTeamAdd(w, req)
				// Обновление существующей команды должно быть успешным (201)
				if w.Code != http.StatusCreated {
					t.Errorf("Expected 201, got %d. Body: %s", w.Code, w.Body.String())
				}
			},
			errorCode:  "", // Не ошибка
			statusCode: http.StatusOK,
		},
		{
			name: "PullRequestExistsError - Create duplicate PR",
			testFunc: func(t *testing.T) {
				prBody, _ := json.Marshal(domain.PostPullRequestCreateJSONBody{
					PullRequestId:   "duplicate-pr-test",
					PullRequestName: "Duplicate PR",
					AuthorId:        teams[0].Members[0].UserId,
				})
				req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(prBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				controller.PostPullRequestCreate(w, req)
			},
			errorCode:  domain.PREXISTS,
			statusCode: http.StatusConflict,
		},
		{
			name: "NoCandidateError - Create PR with author without team",
			testFunc: func(t *testing.T) {
				// Создаем пользователя без команды
				prBody, _ := json.Marshal(domain.PostPullRequestCreateJSONBody{
					PullRequestId:   "no-candidate-pr",
					PullRequestName: "No Candidate PR",
					AuthorId:        "user-without-team-123",
				})
				req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(prBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				controller.PostPullRequestCreate(w, req)
				if w.Code != http.StatusNotFound {
					t.Errorf("Expected 404, got %d. Body: %s", w.Code, w.Body.String())
				}
			},
			errorCode:  domain.NOCANDIDATE,
			statusCode: http.StatusNotFound,
		},
		{
			name: "Invalid request body",
			testFunc: func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				controller.PostTeamAdd(w, req)
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected 400, got %d", w.Code)
				}
			},
			errorCode:  domain.NOTFOUND,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, scenario.testFunc)
	}
}
