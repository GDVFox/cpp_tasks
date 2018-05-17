(use-syntax (ice-9 syncase))

;Рекурсивный макрос для создания get- функций полей структуры
(define-syntax get-fields
  (syntax-rules ()
    ((_ name (field-name-1))
     (eval `(define (,(string->symbol (string-append (symbol->string 'name) "-" (symbol->string 'field-name-1))) object)
              (cdr (assq 'field-name-1 object))) (interaction-environment)))
    ((_ name (field-name-1 field-name-2* ...))
     (begin (eval `(define (,(string->symbol (string-append (symbol->string 'name) "-" (symbol->string 'field-name-1))) object)
                     (cdr (assq 'field-name-1 object))) (interaction-environment))
            (get-fields name (field-name-2* ...))))))

;Рекурсивный макрос для создания set- функций полей структуры
(define-syntax set-fields
  (syntax-rules ()
    ((_ name (field-name-1))
     (eval `(define (,(string->symbol (string-append "set-" (symbol->string 'name) "-" (symbol->string 'field-name-1) "!")) object val)
              (set-cdr! (assq 'field-name-1 object) val)) (interaction-environment)))
    ((_ name (field-name-1 field-name-2* ...))
     (begin (eval `(define (,(string->symbol (string-append "set-" (symbol->string 'name) "-" (symbol->string 'field-name-1) "!")) object val)
                     (set-cdr! (assq 'field-name-1 object) val)) (interaction-environment))
            (set-fields name (field-name-2* ...))))))

;Создание конструктора и предиката + set- и get- функций
(define-syntax define-struct
  (syntax-rules ()
    ((_ name (field-name-1 field-name-2* ...))
     (begin (get-fields name (field-name-1 field-name-2* ...)) ;get- функции
            (set-fields name (field-name-1 field-name-2* ...)) ;set- функции
            (eval `(define (,(string->symbol (string-append "make-" (symbol->string 'name))) . field-vals)
                     (cons (cons 'type 'name)
                           (map (lambda (field-name field-val) (cons field-name field-val))
                                '(field-name-1 field-name-2* ...) field-vals))) (interaction-environment)) ;Конструктор
            (eval `(define (,(string->symbol (string-append (symbol->string 'name) "?")) object)
                     (and (not (null? object)) (list? object) (eq? (cdr (car object)) 'name))) (interaction-environment)))))) ;Предикат