# Fees

Currently DN supports transactions only with non-zero fees in dfi cryptocurrency, so it means each transaction
must contains at least **100000000000000dfi** (dfi has 18 decimals, so that fee could be interpreted as 0.0001 dfi).

So current default fees in **dncli** are **100000000000000dfi**, you can ignore **--fees** flag if you want to send transaction with default amount. 
