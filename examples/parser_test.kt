$var]tekst
$tab.eee[c][d("$if a: eee $end")][a[1]]
hihi$fun(a, "sss", f(s))ooo

$:{fun1(d).fun2(a).fun1(c).wynik}

To jest teskst
wieloliniowy!
$:{vvv}
$fun(argument)
$Zrob(   "$if a: eee $elif b: pp $end" ,a2,a3, a4 ,a5)

$Pokaz("text")
$ToTo("txt $var:ddd $" a")


$:Fajna('
  txt ${var}ccc $'
  $for i,v in tab:
	<h1>Acha $i, to jest $v</h1>
  $end
  $if f("aaa $if a: ccc $elif n: eee $end"):
    Pierwszy blok
  $elif c:
    Drugi blok
  $elif d:
    Trzeci blok
  $else:
    Czwarty blok
  $end
')
$:zmienna

$$
$'
$"

$if var1:
 fffddeeff
  aaaaaa
$elif fun("ddadd"):sssssss $else:
  ddddddd
$end

$for i, v in fun("ddd $ppp eee"):
  aaaa sss
$else:
  bbbb sss
${end}

$for a+ ,b in c:
 <p>$a: $b</p>
$end
