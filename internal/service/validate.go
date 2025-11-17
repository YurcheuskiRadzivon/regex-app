package service

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyPattern = errors.New("пустое регулярное выражение")
)

type validator struct{}

func (v *validator) ValidatePattern(pattern string) error {
	if pattern == "" {
		return ErrEmptyPattern
	}

	runes := []rune(pattern)

	var stack []rune

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if ch == '\\' {
			if i == len(runes)-1 {
				return fmt.Errorf("экранирование в конце строки")
			}
			i++
			continue
		}

		switch ch {
		case '(', '[', '{':
			stack = append(stack, ch)
		case ')', ']', '}':
			if len(stack) == 0 {
				return fmt.Errorf("лишняя закрывающая скобка %q в позиции %d", ch, i)
			}
			open := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if !matchingBrackets(open, ch) {
				return fmt.Errorf("несоответствующие скобки %q и %q", open, ch)
			}
		}
	}

	if len(stack) > 0 {
		return fmt.Errorf("незакрытые скобки")
	}

	for i, ch := range runes {
		if ch == '*' || ch == '+' || ch == '?' {
			if i == 0 {
				return fmt.Errorf("квантификатор %q не может быть в начале выражения", ch)
			}
			prev := runes[i-1]
			if prev == '|' || prev == '(' {
				return fmt.Errorf("квантификатор %q должен следовать за символом/группой", ch)
			}
		}
	}

	if err := v.validateCharClasses(runes); err != nil {
		return err
	}

	return nil
}

func (v *validator) validateCharClasses(runes []rune) error {
	for i := 0; i < len(runes); i++ {
		if runes[i] == '\\' {
			i++
			continue
		}
		if runes[i] == '[' {
			i++
			if i < len(runes) && runes[i] == '^' {
				i++
			}
			var prev rune
			hasPrev := false
			for ; i < len(runes) && runes[i] != ']'; i++ {
				ch := runes[i]
				if ch == '\\' {
					if i == len(runes)-1 {
						return fmt.Errorf("неполное экранирование в классе символов")
					}
					i++
					ch = runes[i]
				}
				if hasPrev && ch == '-' && i+1 < len(runes) && runes[i+1] != ']' {
					next := runes[i+1]
					if next == '\\' {
						if i+2 >= len(runes) {
							return fmt.Errorf("неполный диапазон в классе символов")
						}
						next = runes[i+2]
					}
					if prev > next {
						return fmt.Errorf("некорректный диапазон %q-%q в классе символов", prev, next)
					}
				}
				prev = ch
				hasPrev = true
			}
			if i >= len(runes) || runes[i] != ']' {
				return fmt.Errorf("незакрытый класс символов []")
			}
		}
	}
	return nil
}

func matchingBrackets(open, close rune) bool {
	return (open == '(' && close == ')') ||
		(open == '[' && close == ']') ||
		(open == '{' && close == '}')
}
