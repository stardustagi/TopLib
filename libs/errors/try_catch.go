package errors

// Try 模拟 try/catch/finally 结构
func TryFunc(fun func(), catch func(err interface{}), finally func()) {
	defer func() {
		// finally 块无论如何都会执行
		if finally != nil {
			finally()
		}

		// 如果捕获到 panic，执行 catch
		if r := recover(); r != nil {
			if catch != nil {
				catch(r)
			}
		}
	}()

	// 执行 try 块
	fun()
}
