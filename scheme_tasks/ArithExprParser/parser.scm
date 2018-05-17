;; Грамматика в БНФ для tokenize
;; <число> ::= [<знак>]<число_без_знака>
;; <число_без_знака> ::= <цифра>|<число_без_знака><цифра>
;; <знак> ::= +|-
;; <цифра> ::= 0|1|2|3|4|5|6|7|8|9
;; <идентификатор> ::= <буква>|<идентификатор><буква>
;; <буква> ::= a|b|c|d|...|x|y|z
;; <откр_скобка> ::= (
;; <закр_скобка> ::= )

(define (string-integer? str)
  (string->number str))

(define (normal-ident? str)
  (or (null? str) (and (char-alphabetic? (car str)) (normal-ident? (cdr str)))))

(define (parenthesis? str)
  (or (equal? str "(") (equal? str ")")))

(define (op? ch)
  (or (eq? ch #\+) (eq? ch #\-) (eq? ch #\*) (eq? ch #\/) (eq? ch #\^)))

(define (tokenize str)
  (define (check-end-of-lexema str start-indx end-indx value substr result) ;; Проверяет закончилась ли лексема
    (if (not (= (+ start-indx 1) end-indx))                                 ;; переводит лексему в токен и сохраняет, продолжает рекурсию
        (or (and (string-integer? value) (tokenize-in str end-indx end-indx (cons substr (cons (string->number value) result))))
            (and (normal-ident? (string->list value)) (tokenize-in str end-indx end-indx (cons substr (cons (string->symbol value) result)))))
        (tokenize-in str end-indx end-indx (cons substr result))))
  (define (tokenize-in str start-indx end-indx result)
    (cond ((>= end-indx (string-length str)) ;; Действия, когда строка закончилась
           (if (not (= start-indx end-indx))
               (let ((value (substring str start-indx end-indx)))
                 (or (and (string-integer? value) (cons (string->number value) result))
                     (and (normal-ident? (string->list value)) (cons (string->symbol value) result))))
               result))
          ((and (op? (string-ref str end-indx)) (or (= end-indx 0) (not (eq? (char-downcase (string-ref str (- end-indx 1))) #\e)))) ;; Если текущий символ - оператор
           (check-end-of-lexema str start-indx (+ end-indx 1) (substring str start-indx end-indx)
                                (string->symbol (substring str end-indx (+ end-indx 1))) result))
          ((parenthesis? (substring str end-indx (+ end-indx 1))) ;; Если текущий символ - скобка
           (check-end-of-lexema str start-indx (+ end-indx 1) (substring str start-indx end-indx)
                                (substring str end-indx (+ end-indx 1)) result))
          ((char-whitespace? (string-ref str end-indx)) ;; Если текущий символ - пробел
           (if (not (= start-indx end-indx))
               (let ((value (substring str start-indx end-indx)))
                 (or (and (string-integer? value) (tokenize-in str (+ end-indx 1) (+ end-indx 1) (cons (string->number value) result)))
                     (and (normal-ident? (string->list value)) (tokenize-in str (+ end-indx 1) (+ end-indx 1) (cons (string->symbol value) result)))))
               (tokenize-in str (+ end-indx 1) (+ end-indx 1) result)))
          (else (tokenize-in str start-indx (+ end-indx 1) result))))
  (let ((result (tokenize-in str 0 0 '())))
    (and result (reverse result)))) ;; tokenize-in возвращает список в обратном порядке

(define (parse lst)
  (define (list-head xs k) ;; голова списка до k-ого элемента (k не включая)
    (reverse (list-tail (reverse xs) (- (length xs) k))))
  (define (find-op lst answ ops br-counter) ;; Поиск одного из двух операторов из ops в списке lst.
    (if (null? lst)                         ;; Пропускает операторы внутри скобок
        'no-op                              ;; br-counter считает вложенность
        (cond ((equal? (car lst) ")") (find-op (cdr lst) (+ answ 1) ops (+ br-counter 1)))
              ((equal? (car lst) "(") (find-op (cdr lst) (+ answ 1) ops (- br-counter 1)))
              ((and (= br-counter 0) (eq? (car lst) (car ops))) (cons answ (car ops)))
              ((and (= br-counter 0) (eq? (car lst) (cdr ops))) (cons answ (cdr ops)))
              (else (find-op (cdr lst) (+ answ 1) ops br-counter)))))
  (define (get-expr lst result br-counter)                                                       ;; Возвращает список с выражением до первой закрывающей скобки
    (cond ((null? lst) #f) ;; ошбика, нет закрывающей скобки
          ((equal? (car lst) "(") (get-expr (cdr lst) (cons (car lst) result) (+ br-counter 1))) ;; br-counter считает вложенность
          ((equal? (car lst) ")") (if (= br-counter 0) (reverse result) (get-expr (cdr lst) (cons (car lst) result) (- br-counter 1))))
          (else (get-expr (cdr lst) (cons (car lst) result) br-counter))))
  (define (term lst) ;; Cоответствует Term из БНФ
    (let ((mul-div-indx (find-op (reverse lst) 0 '(* . /) 0)))  ;; поиск индекса первого справа (для левой ассоц) знака умножения или деления
      (if (not (pair? mul-div-indx))
          (factor lst)
          (list (term (list-head lst (- (length lst) (car mul-div-indx) 1)))
                (cdr mul-div-indx) (factor (list-tail lst (- (length lst) (car mul-div-indx))))))))
  (define (factor lst) ;; Соответствует Factor из БНФ
    (let ((pow-indx (find-op lst 0 '(^ . ^) 0))) ;; поиск индекса первого слева (для правой ассоц) знака возведения в степень
      (if (not (pair? pow-indx))
          (power lst)
          (list (power (list-head lst (car pow-indx)))
                (cdr pow-indx) (factor (list-tail lst (+ (car pow-indx) 1)))))))
  (define (power lst) ;; Соответсвует Power из БНФ
    (cond ((null? lst) #f)
          ((eq? (car lst) '-) (list '- (cadr lst)))
          ((equal? (car lst) "(") (expr (get-expr (cdr lst) '() 0)))
          (else (and (= (length lst) 1) (car lst))))) ;; ошибка, нет оператора между операндами
  (define (expr lst) ;; Соответсвует Expr из БНФ
    (and lst
         (let ((plus-munis-indx (find-op (reverse lst) 0 '(+ . -) 0))) ;; поиск индекса первого справа (для левой ассоц) знака сложения или вычитания
           (if (not (pair? plus-munis-indx))
               (term lst)
               (if (= (- (length lst) (car plus-munis-indx) 1) 0)
                   (list (cdr plus-munis-indx) (term (list-tail lst (- (length lst) (car plus-munis-indx)))))
                   (list (expr (list-head lst (- (length lst) (car plus-munis-indx) 1)))
                         (cdr plus-munis-indx) (term (list-tail lst (- (length lst) (car plus-munis-indx))))))))))
  (define (check-result res) ;; Проверка: были ли ошибки во время парсинга
    (or (null? res) (and (if (list? (car res)) (check-result (car res)) (car res)) (check-result (cdr res)))))
  (and lst (let ((result (expr lst)))
             (or (and (not (list? result)) result) (and result (check-result result) result)))))
    
(define (tree->scheme tree)
  (if (not (list? tree))
      tree
      (if (>= (length tree) 3)
          `(,(if (eq? (cadr tree) '^) 'expt (cadr tree)) ,(tree->scheme (car tree)) ,(tree->scheme (caddr tree)))
          `(- ,(tree->scheme (cadr tree))))))