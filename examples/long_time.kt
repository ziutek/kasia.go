START

$F(0)("$$$A $"and$" $$$["A"]", 1)
$['F'](0)('$$$A "and" $$$['A']', 1)

$for i, v in D:
  $i: $if i == 1:
    $v(1.1)
  $else:
    D[$i] = $v
  $end
$else:
  D == nil or len(D) == 0
$end

$A    ==  $[0]
$B    ==  $[1]
$C    ==  $[2]
$D[A] ==  $[3][A]   // $A is integer used as index
$:E.a  ==  $:[4].a
$:E.b  ==  $:[4].b
$G.A  ==  $[6][0]

$:M1
$M1()
$G.M2("A")

eee${A}eee uuu$:{A}uuu

$# Comment #$

END
