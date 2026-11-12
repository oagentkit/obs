package langfuse

// 本文件钉 oagentkit/obs#105 的回归：FetchBatch 的 detail 阶段整体 ctx 取消
// 必须上抛，不得并进「单条失败降级」档返回 err==nil 的空批次。此前 derr 只被
// 当计数器用（errors.As 取 Failed 条数），context.Canceled 既不上抛也不判别
// ——同一个 ctx 取消在同一个函数里得到相反处置（列表阶段上抛、detail 阶段
// 静默），调用方无法分辨「trace 被删了」与「我被取消了」。

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// TestFetchBatchDetailPhaseCtxCancelPropagates 实验组 A：detail 阶段整体
// ctx 取消 ⇒ err 必须非 nil 且可 errors.Is 命中 context.Canceled——
// 与列表阶段（对照 D）同口径。
func TestFetchBatchDetailPhaseCtxCancelPropagates(t *testing.T) {
	detailStarted := make(chan struct{}, 1)
	release := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/traces", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": 1, "limit": 50, "totalItems": 2, "totalPages": 1},
			"data": []map[string]any{
				{"id": "tr-1", "timestamp": "2026-09-14T08:00:00Z", "name": "agent"},
				{"id": "tr-2", "timestamp": "2026-09-14T08:01:00Z", "name": "agent"},
			},
		})
	})
	mux.HandleFunc("/api/public/traces/", func(w http.ResponseWriter, r *http.Request) {
		// 详情端点阻塞到测试收尾——确保取消发生在 detail 阶段（列表已成功）。
		select {
		case detailStarted <- struct{}{}:
		default:
		}
		<-release
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": r.URL.Path[len("/api/public/traces/"):], "timestamp": "2026-09-14T08:00:00Z",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close() // LIFO：先放行 handler，再关 server（否则 Close 挂住）
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-detailStarted // 列表已收齐、详情已开跑 ⇒ 取消落在 detail 阶段
		cancel()
	}()

	batch, err := NewClient(srv.URL, "pk", "sk").FetchBatch(ctx, Query{Limit: 2})
	if err == nil {
		t.Fatal("detail 阶段整体 ctx 取消必须上抛 err（obs#105）：err==nil 的空批次与「单条 404 降级」同形，调用方无从分辨")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("上抛的错误应可 errors.Is 命中 context.Canceled，实得: %v", err)
	}
	if len(batch.Traces) != 0 {
		t.Fatalf("取消路径不该有完整 traces: %+v", batch)
	}
}

// TestFetchBatchDetailPhaseAll404Degrades 对照组 B：detail 全部 404（契约
// 承诺的降级路径）仍返回 err==nil + Skipped 计数——上修不得把这条档位收掉。
func TestFetchBatchDetailPhaseAll404Degrades(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/traces", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": 1, "limit": 50, "totalItems": 2, "totalPages": 1},
			"data": []map[string]any{
				{"id": "tr-1", "timestamp": "2026-09-14T08:00:00Z", "name": "agent"},
				{"id": "tr-2", "timestamp": "2026-09-14T08:01:00Z", "name": "agent"},
			},
		})
	})
	mux.HandleFunc("/api/public/traces/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"trace not found"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	batch, err := NewClient(srv.URL, "pk", "sk").FetchBatch(context.Background(), Query{Limit: 2})
	if err != nil {
		t.Fatalf("单条 404 降级档不得被上修收掉（obs#105 对照面）: %v", err)
	}
	if batch.Skipped != 2 || len(batch.Traces) != 0 {
		t.Fatalf("404 应计 Skipped 不弃语义: %+v", batch)
	}
}

// TestFetchBatchListPhaseCtxCancelPropagates 对照组 D：同一个 ctx 取消落在
// 列表阶段本来就上抛——两阶段口径必须一致（#105 修复后 detail 阶段对齐此面）。
func TestFetchBatchListPhaseCtxCancelPropagates(t *testing.T) {
	var listCalls atomic.Int64
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/traces", func(w http.ResponseWriter, r *http.Request) {
		listCalls.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"meta": map[string]any{"page": 1, "limit": 50, "totalItems": 1, "totalPages": 1},
			"data": []map[string]any{{"id": "tr-1", "timestamp": "2026-09-14T08:00:00Z"}},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 列表阶段前已取消
	batch, err := NewClient(srv.URL, "pk", "sk").FetchBatch(ctx, Query{Limit: 1})
	if err == nil {
		t.Fatal("列表阶段 ctx 取消必须上抛（既有语义，钉住与 detail 阶段一致）")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("应命中 context.Canceled，实得: %v", err)
	}
	if listCalls.Load() != 0 {
		t.Fatalf("已取消的 ctx 不该发起列表请求，实发 %d 次", listCalls.Load())
	}
	if len(batch.Traces) != 0 {
		t.Fatalf("取消路径不该有 traces: %+v", batch)
	}
}
