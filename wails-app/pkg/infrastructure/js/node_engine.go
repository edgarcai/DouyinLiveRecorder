package js

import (
    "bytes"
    "os/exec"
)

// NodeEngine 使用本地 Node 执行指定脚本并返回结果，用于签名计算等场景。
type NodeEngine struct{}

// Eval 执行给定脚本路径（如 src/javascript/x-bogus.js），返回 stdout 作为结果。
func (n *NodeEngine) Eval(script string) (string, error) {
    cmd := exec.Command("node", script)
    var out bytes.Buffer
    var errb bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &errb
    if err := cmd.Run(); err != nil { return "", err }
    return out.String(), nil
}

// EvalWithArgs 传入脚本与参数（如待签名查询字符串），返回 stdout。
func (n *NodeEngine) EvalWithArgs(script string, args []string) (string, error) {
    // 以 -e 方式执行：加载模块并打印签名结果（约定前两个参数为 script 与第一个业务参数）
    code := "const m=require(process.argv[1]);const arg1=process.argv[2]||'';const arg2=process.argv[3]||'';if(m.sign){console.log(m.sign(arg1,arg2));}else{console.log('')}}"
    argv := append([]string{"-e", code, script}, args...)
    cmd := exec.Command("node", argv...)
    var out bytes.Buffer
    var errb bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &errb
    if err := cmd.Run(); err != nil { return "", err }
    return out.String(), nil
}
