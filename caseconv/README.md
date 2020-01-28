# caseconv

case provides functions to split or join string in different styles.

## example
```
package main

import (
	"fmt"
	"github.com/liguangsheng/tile/caser"
)

func main() {
	fmt.Println(caser.CamelSplit("HTTPurlID"))       // [HTTP url ID]
	fmt.Println(caser.CamelSplit("ThankYou"))        // [Thank You]
	fmt.Println(caser.SnakeSplit("how_are_you"))     // [how are you]
	fmt.Println(caser.PascalSplit("ILoveYou"))       // [I Love You]
	fmt.Println(caser.KebabSplit("how-old-are-you")) // [how old are you]

	parts := []string{"i", "love", "you"}
	fmt.Println(caser.UpperCamelJoin(parts)) // ILoveYou
	fmt.Println(caser.UpperKebabJoin(parts)) // I-LOVE-YOU
	fmt.Println(caser.UpperSnakeJoin(parts)) // I_LOVE_YOU
	fmt.Println(caser.LowerCamelJoin(parts)) // iLoveYou
	fmt.Println(caser.LowerKebabJoin(parts)) // i-love-you
	fmt.Println(caser.LowerSnakeJoin(parts)) // i_love_you
	fmt.Println(caser.PascalJoin(parts))     // ILoveYou
	fmt.Println(caser.TitleKebabJoin(parts)) // I-Love-You
	fmt.Println(caser.TitleSnakeJoin(parts)) // I_Love_You
	fmt.Println(caser.PascalJoin(parts))     // ILoveYou
}
```