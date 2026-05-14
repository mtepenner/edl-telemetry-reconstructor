/*
Simple matrix operations for EKF estimation
Uses pure Go without external dependencies for reliability
*/

package estimation

import (
	"math"
)

// Matrix holds a dense matrix for EKF operations
type Matrix struct {
	data [][]float64
	rows int
	cols int
}

// Vector is a column vector
type Vector struct {
	data []float64
	n    int
}

// NewMatrix creates a new matrix
func NewMatrix(r, c int) *Matrix {
	data := make([][]float64, r)
	for i := 0; i < r; i++ {
		data[i] = make([]float64, c)
	}
	return &Matrix{data: data, rows: r, cols: c}
}

// NewVector creates a new column vector
func NewVector(n int) *Vector {
	return &Vector{data: make([]float64, n), n: n}
}

// NewVectorFromSlice creates a vector from a slice
func NewVectorFromSlice(data []float64) *Vector {
	v := NewVector(len(data))
	copy(v.data, data)
	return v
}

// ToSlice converts vector to slice
func (v *Vector) ToSlice() []float64 {
	result := make([]float64, v.n)
	copy(result, v.data)
	return result
}

// Set sets a value at position (i, j)
func (m *Matrix) Set(i, j int, v float64) {
	if i >= 0 && i < m.rows && j >= 0 && j < m.cols {
		m.data[i][j] = v
	}
}

// SetVec sets a vector element (for column vectors)
func (v *Vector) SetVec(i int, val float64) {
	if i >= 0 && i < v.n {
		v.data[i] = val
	}
}

// Get gets a value at position (i, j)
func (m *Matrix) Get(i, j int) float64 {
	if i >= 0 && i < m.rows && j >= 0 && j < m.cols {
		return m.data[i][j]
	}
	return 0
}

// At gets a value (same as Get)
func (m *Matrix) At(i, j int) float64 {
	return m.Get(i, j)
}

// Rows returns number of rows
func (m *Matrix) Rows() int {
	return m.rows
}

// Cols returns number of columns
func (m *Matrix) Cols() int {
	return m.cols
}

// Vector At
func (v *Vector) At(i, j int) float64 {
	if i >= 0 && i < v.n && j == 0 {
		return v.data[i]
	}
	return 0
}

// Vector Rows
func (v *Vector) Rows() int {
	return v.n
}

// Vector Cols
func (v *Vector) Cols() int {
	return 1
}

// Multiply multiplies two matrices: result = a * b
func Multiply(a, b *Matrix) *Matrix {
	result := NewMatrix(a.rows, b.cols)
	for i := 0; i < a.rows; i++ {
		for j := 0; j < b.cols; j++ {
			sum := 0.0
			for k := 0; k < a.cols; k++ {
				sum += a.Get(i, k) * b.Get(k, j)
			}
			result.Set(i, j, sum)
		}
	}
	return result
}

// MultiplyVector multiplies matrix by vector: result = a * v
func MultiplyVector(a *Matrix, v *Vector) *Vector {
	result := NewVector(a.rows)
	for i := 0; i < a.rows; i++ {
		sum := 0.0
		for j := 0; j < a.cols; j++ {
			sum += a.Get(i, j) * v.data[j]
		}
		result.data[i] = sum
	}
	return result
}

// Transpose returns transposed matrix
func Transpose(m *Matrix) *Matrix {
	result := NewMatrix(m.cols, m.rows)
	for i := 0; i < m.rows; i++ {
		for j := 0; j < m.cols; j++ {
			result.Set(j, i, m.Get(i, j))
		}
	}
	return result
}

// Add adds two matrices: result = a + b
func Add(a, b *Matrix) *Matrix {
	result := NewMatrix(a.rows, a.cols)
	for i := 0; i < a.rows; i++ {
		for j := 0; j < a.cols; j++ {
			result.Set(i, j, a.Get(i, j)+b.Get(i, j))
		}
	}
	return result
}

// Subtract subtracts matrices: result = a - b
func Subtract(a, b *Matrix) *Matrix {
	result := NewMatrix(a.rows, a.cols)
	for i := 0; i < a.rows; i++ {
		for j := 0; j < a.cols; j++ {
			result.Set(i, j, a.Get(i, j)-b.Get(i, j))
		}
	}
	return result
}

// Identity creates an identity matrix
func Identity(n int) *Matrix {
	result := NewMatrix(n, n)
	for i := 0; i < n; i++ {
		result.Set(i, i, 1.0)
	}
	return result
}

// VectorNorm computes the Euclidean norm of a vector
func VectorNorm(v *Vector) float64 {
	sum := 0.0
	for i := 0; i < v.n; i++ {
		sum += v.data[i] * v.data[i]
	}
	return math.Sqrt(sum)
}
