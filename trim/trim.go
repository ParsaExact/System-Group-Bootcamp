package trim

func TrimSpace(s string) string {
    ans := ""
    for i := 0; i < len(s); i++ {
        if s[i] != ' ' {
            ans = s[i:]
            break
        }
    }
    for i := len(ans) - 1; i >= 0; i-- {
        if ans[i] != ' ' {
            ans = ans[:i+1]
            break
        }
    }
    return ans
}