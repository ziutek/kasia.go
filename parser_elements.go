package kasia

type Element interface {
    lnum() int
}

type TxtElem struct {
    ln  int
    txt []byte
}
func (self *TxtElem) lnum() int { return self.ln }

type VarFunElem struct {
    ln   int
    name interface{}
    filt bool
    fun  bool
    args []interface{}
    next *VarFunElem
}
func (self *VarFunElem) lnum() int { return self.ln }

const (
    if_nocmp = iota // if jednoargumentowy
    if_eq           // rowny
    if_ne           // nierowy
    if_lt           // mniejsz
    if_le           // mniejszy lub rowny
    if_gt           // wiekszy
    if_ge           // wiekszy lub rowny
)

type IfElem struct {
    ln          int
    cmp         int
    arg1        interface{}
    arg2        interface{}
    true_block  []Element
    false_block []Element
}
func (self *IfElem) lnum() int { return self.ln }

type ElifElem struct {
    ln   int
    cmp  int
    arg1 interface{}
    arg2 interface{}
}
func (self *ElifElem) lnum() int { return self.ln }

type ElseElem struct {
    ln int
}
func (self *ElseElem) lnum() int { return self.ln }

type EndElem struct {
    ln int
}
func (self *EndElem) lnum() int { return self.ln }

type ForElem struct {
    ln         int
    iter_inc   int
    iter       string
    val        string
    arg        *VarFunElem
    iter_block []Element
    else_block []Element
}
func (self *ForElem) lnum() int { return self.ln }
