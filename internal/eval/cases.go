package eval

// Case 评测用例
type Case struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Input     string `json:"input"`
}

// DefaultCases 题目要求的 2 条 + 可选扩展
var DefaultCases = []Case{
	{ID: "1", Direction: "product_to_dev", Input: "我们需要一个智能推荐功能，提升用户停留时长"},
	{ID: "2", Direction: "dev_to_product", Input: "我们优化了数据库查询，QPS 提升了 30%"},
}

// Result 单条评测结果
type Result struct {
	CaseID     string `json:"case_id"`
	Pass       bool   `json:"pass"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"` // 模型/Agent 失败时的完整错误链
	DurationMs int64  `json:"duration_ms"`
}

// RunResult 一次评测运行结果
type RunResult struct {
	Results    []Result `json:"results"`
	DurationMs int64    `json:"duration_ms"`
}
