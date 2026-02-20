package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/MorantHP/OURERP/internal/models"
	"github.com/MorantHP/OURERP/internal/repository"
)

// SchedulerService 调度服务
type SchedulerService struct {
	syncService *SyncService
	shopRepo    *repository.ShopRepository
	tickers     map[int64]*time.Ticker
	mu          sync.RWMutex
	stopChan    chan struct{}
	running     bool
}

// NewSchedulerService 创建调度服务
func NewSchedulerService(syncService *SyncService, shopRepo *repository.ShopRepository) *SchedulerService {
	return &SchedulerService{
		syncService: syncService,
		shopRepo:    shopRepo,
		tickers:     make(map[int64]*time.Ticker),
		stopChan:    make(chan struct{}),
	}
}

// Start 启动调度服务
func (s *SchedulerService) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	log.Println("调度服务已启动")

	// 加载所有启用的店铺并设置定时任务
	status := models.ShopStatusEnabled
	shops, _, err := s.shopRepo.List(1, 1000, "", &status)
	if err != nil {
		log.Printf("加载店铺列表失败: %v", err)
		return err
	}

	for _, shop := range shops {
		if shop.SyncInterval > 0 {
			s.scheduleShop(shop.ID, shop.SyncInterval)
		}
	}

	return nil
}

// Stop 停止调度服务
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)

	for shopID, ticker := range s.tickers {
		ticker.Stop()
		delete(s.tickers, shopID)
	}

	s.running = false
	log.Println("调度服务已停止")
}

// scheduleShop 为店铺设置定时同步
func (s *SchedulerService) scheduleShop(shopID int64, intervalMinutes int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 如果已存在，先停止
	if ticker, exists := s.tickers[shopID]; exists {
		ticker.Stop()
	}

	interval := time.Duration(intervalMinutes) * time.Minute
	ticker := time.NewTicker(interval)
	s.tickers[shopID] = ticker

	go func() {
		for {
			select {
			case <-ticker.C:
				s.syncShop(shopID)
			case <-s.stopChan:
				return
			}
		}
	}()

	log.Printf("店铺 [%d] 定时同步已设置，间隔: %d分钟", shopID, intervalMinutes)
}

// SyncShopNow 立即同步店铺
func (s *SchedulerService) SyncShopNow(shopID int64) (*SyncResult, error) {
	return s.syncShop(shopID)
}

// syncShop 同步店铺订单
func (s *SchedulerService) syncShop(shopID int64) (*SyncResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 获取最近24小时的订单
	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)

	result, err := s.syncService.SyncShopOrders(ctx, shopID, startTime, endTime)
	if err != nil {
		log.Printf("店铺 [%d] 同步失败: %v", shopID, err)
		return result, err
	}

	log.Printf("店铺 [%d] 同步完成: 成功 %d, 失败 %d, 耗时 %v",
		shopID, result.TotalSynced, result.TotalFailed, result.Duration)

	return result, nil
}

// UpdateShopInterval 更新店铺同步间隔
func (s *SchedulerService) UpdateShopInterval(shopID int64, intervalMinutes int) {
	s.scheduleShop(shopID, intervalMinutes)
}

// RemoveShop 移除店铺定时任务
func (s *SchedulerService) RemoveShop(shopID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ticker, exists := s.tickers[shopID]; exists {
		ticker.Stop()
		delete(s.tickers, shopID)
		log.Printf("店铺 [%d] 定时同步已移除", shopID)
	}
}

// SyncAllShops 同步所有启用的店铺
func (s *SchedulerService) SyncAllShops() map[int64]*SyncResult {
	results := make(map[int64]*SyncResult)

	status := models.ShopStatusEnabled
	shops, _, err := s.shopRepo.List(1, 1000, "", &status)
	if err != nil {
		log.Printf("加载店铺列表失败: %v", err)
		return results
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, shop := range shops {
		wg.Add(1)
		go func(shopID int64) {
			defer wg.Done()
			result, err := s.syncShop(shopID)
			mu.Lock()
			if err != nil {
				results[shopID] = &SyncResult{
					ShopID:       shopID,
					ErrorMessage: err.Error(),
				}
			} else {
				results[shopID] = result
			}
			mu.Unlock()
		}(shop.ID)
	}

	wg.Wait()
	return results
}

// GetRunningTasks 获取运行中的任务
func (s *SchedulerService) GetRunningTasks() []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]int64, 0, len(s.tickers))
	for shopID := range s.tickers {
		tasks = append(tasks, shopID)
	}
	return tasks
}
