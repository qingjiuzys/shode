package benchmarks

import (
	"testing"

	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func BenchmarkStdLib_StringOperations(b *testing.B) {
	sl := stdlib.New()
	
	b.Run("ToUpper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.ToUpper("hello world")
		}
	})
	
	b.Run("ToLower", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.ToLower("HELLO WORLD")
		}
	})
	
	b.Run("Trim", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.Trim("  hello world  ")
		}
	})
	
	b.Run("Replace", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.Replace("hello world", "world", "there")
		}
	})
}

func BenchmarkStdLib_EnvOperations(b *testing.B) {
	sl := stdlib.New()
	sl.SetEnv("TEST", "value")
	
	b.Run("GetEnv", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.GetEnv("TEST")
		}
	})
	
	b.Run("SetEnv", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.SetEnv("KEY", "value")
		}
	})
}

func BenchmarkStdLib_CacheOperations(b *testing.B) {
	sl := stdlib.New()
	
	b.Run("SetCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.SetCache("key", "value", 60)
		}
	})
	
	b.Run("GetCache", func(b *testing.B) {
		sl.SetCache("key", "value", 60)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sl.GetCache("key")
		}
	})
	
	b.Run("DeleteCache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.SetCache("key", "value", 60)
			sl.DeleteCache("key")
		}
	})
}

func BenchmarkStdLib_FileOperations(b *testing.B) {
	sl := stdlib.New()
	
	b.Run("FileExists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.FileExists("/tmp")
		}
	})
	
	b.Run("WorkingDir", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.WorkingDir()
		}
	})
}

func BenchmarkStdLib_HashOperations(b *testing.B) {
	sl := stdlib.New()
	
	b.Run("SHA256Hash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sl.SHA256Hash("test string for hashing")
		}
	})
}
