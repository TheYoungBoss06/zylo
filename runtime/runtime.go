package zyloruntime

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

// Println es una función de runtime para imprimir en la consola.
// Corresponde a la función 'print' en Zylo.
func Println(a ...interface{}) {
	fmt.Println(a...)
}

// Print es una función de runtime para imprimir sin nueva línea.
func Print(a ...interface{}) {
	fmt.Print(a...)
}

// Printf es una función de runtime para imprimir con formato.
func Printf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
}

// Exit es una función de runtime para terminar el programa.
func Exit(code int) {
	os.Exit(code)
}

// --- Manejo de Errores y Excepciones ---

// ZyloError representa un error en Zylo.
type ZyloError struct {
	Message string
}

// Error implementa la interfaz error.
func (e *ZyloError) Error() string {
	return e.Message
}

// Throw lanza una excepción Zylo.
func Throw(message string) {
	panic(&ZyloError{Message: message})
}

// Try ejecuta una función y captura excepciones.
func Try(fn func(), catch func(error)) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(*ZyloError); ok {
				catch(err)
			} else {
				// Re-panic para errores no-Zylo
				panic(r)
			}
		}
	}()
	fn()
}

// --- Estructuras de Datos ---

// List representa una lista dinámica.
type List struct {
	items []interface{}
}

// NewList crea una nueva lista.
func NewList() *List {
	return &List{items: make([]interface{}, 0)}
}

// Append añade un elemento al final de la lista.
func (l *List) Append(item interface{}) {
	l.items = append(l.items, item)
}

// Get obtiene un elemento por índice.
func (l *List) Get(index int) interface{} {
	if index < 0 || index >= len(l.items) {
		Throw("Index out of bounds")
	}
	return l.items[index]
}

// Set establece un elemento en un índice.
func (l *List) Set(index int, item interface{}) {
	if index < 0 || index >= len(l.items) {
		Throw("Index out of bounds")
	}
	l.items[index] = item
}

// Len devuelve la longitud de la lista.
func (l *List) Len() int {
	return len(l.items)
}

// Map representa un mapa de clave-valor.
type Map struct {
	items map[string]interface{}
}

// NewMap crea un nuevo mapa.
func NewMap() *Map {
	return &Map{items: make(map[string]interface{})}
}

// Set establece un valor para una clave.
func (m *Map) Set(key string, value interface{}) {
	m.items[key] = value
}

// Get obtiene un valor por clave.
func (m *Map) Get(key string) interface{} {
	return m.items[key]
}

// Has verifica si una clave existe.
func (m *Map) Has(key string) bool {
	_, exists := m.items[key]
	return exists
}

// Delete elimina una clave.
func (m *Map) Delete(key string) {
	delete(m.items, key)
}

// Keys devuelve todas las claves.
func (m *Map) Keys() []string {
	keys := make([]string, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys
}

// --- I/O y Filesystem ---

// ReadFile lee el contenido completo de un archivo.
func ReadFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		Throw(fmt.Sprintf("Error reading file: %v", err))
	}
	return string(content)
}

// WriteFile escribe contenido a un archivo.
func WriteFile(filename string, content string) {
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		Throw(fmt.Sprintf("Error writing file: %v", err))
	}
}

// FileExists verifica si un archivo existe.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// --- JSON ---

// ToJSON convierte un valor a JSON string.
func ToJSON(value interface{}) string {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		Throw(fmt.Sprintf("Error converting to JSON: %v", err))
	}
	return string(jsonBytes)
}

// FromJSON convierte un JSON string a un valor.
func FromJSON(jsonStr string) interface{} {
	var result interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		Throw(fmt.Sprintf("Error parsing JSON: %v", err))
	}
	return result
}

// --- Time y Utilidades ---

// Now devuelve la hora actual como timestamp.
func Now() int64 {
	return time.Now().Unix()
}

// Sleep pausa la ejecución por milisegundos.
func Sleep(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

// Random devuelve un número aleatorio entre 0 y max-1.
func Random(max int) int {
	return int(time.Now().UnixNano()) % max
}

// --- Funciones Matemáticas ---

// Abs devuelve el valor absoluto de un número.
func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Max devuelve el máximo de dos números.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min devuelve el mínimo de dos números.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- Input/Output ---

// ReadLine lee una línea de entrada del usuario.
func ReadLine() string {
	var input string
	fmt.Scanln(&input)
	return input
}

// ReadInt lee un número entero del usuario.
func ReadInt() int {
	var input int
	fmt.Scanf("%d", &input)
	return input
}

// --- String Utilities ---

// StrLen devuelve la longitud de una cadena.
func StrLen(s string) int {
	return len(s)
}

// StrSplit divide una cadena por un separador.
func StrSplit(s, sep string) []string {
	// Implementación simple
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// StrJoin une una lista de cadenas con un separador.
func StrJoin(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// ToNumber convierte un string a número, retorna nil si falla.
func ToNumber(s string) interface{} {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return f
}

// String convierte cualquier valor a string.
func String(x interface{}) string {
	return fmt.Sprintf("%v", x)
}

// Split divide un string y retorna una lista de substrings.
func Split(s, sep string) *List {
	result := NewList()
	current := ""
	for _, char := range s {
		if string(char) == sep {
			result.Append(current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result.Append(current)
	}
	return result
}

// Add suma dos números.
func Add(a, b float64) float64 {
	return a + b
}

// Subtract resta dos números.
func Subtract(a, b float64) float64 {
	return a - b
}

// Multiply multiplica dos números.
func Multiply(a, b float64) float64 {
	return a * b
}

// Divide divide dos números, lanza error si divisor es cero.
func Divide(a, b float64) float64 {
	if b == 0 {
		Throw("Division by zero")
	}
	return a / b
}
