# Errata

## Chapter 1

  * Page 4: The 5th bullet point says:
  
   > There are 128 distinct states in the composite system. (For each state of system 1, system 2 can be in 7 distinct states, so 2^7=128.)
   
   This is wrong. For every state of the first system, there are 7 possible states of the second system, making it 7x7=49 states. So this quote should be:
   > There are 49 distinct states in the composite system. (For each state of system 1, system 2 can be in 7 distinct states, so 7^2=49.)
   
   In general, for a system with `n` states, `m` copies of that system running in parallel will have n^m distinct states.
