#include <string>
#include <fstream>
#include <map>
#include <sstream>
#include <iostream>
#include <unordered_set>
#include <algorithm>

const static std::string dict_file_path = "./text/count_big.txt";

//Rule структура представляющая словарное правило
struct Rule {
    std::string word;                       //Слово
    int frequency;                          //Частота слова

    //Множество биграм слова. Используем hash-table,
    // т.к. потребуется много раз проверять наличие элемента во множестве;
    std::unordered_set<std::string> bigrams;
};

// find_bigrams принимает на вход
// const std::string &word и складывает его биграмы
// во множество std::unordered_set<std::string> &bigrams
void find_bigrams(const std::string &word,
                  std::unordered_set<std::string> &bigrams) {
    if (word.length() == 1) {
        bigrams.insert(word);
        return;
    }

    std::string even_bigram(2,' '), odd_bigram(2,' ');
    even_bigram[0] = word[0];
    for (int i = 1; i < word.length(); i++) {
        even_bigram[i%2] = word[i];
        odd_bigram[(i-1)%2] = word[i];
        bigrams.insert((i % 2 ? even_bigram : odd_bigram));
    }
}

//init_dict инициализирует словарь std::map<std::string, Rule *> &dict,
//парами из файла const std::string &dict_file_path;
void init_dict(const std::string &dict_file_path,
               std::map<std::string, Rule *> &dict) {
    std::ifstream dict_file;
    dict_file.open(dict_file_path);
    std::string line;
    while(std::getline(dict_file, line)) {
        std::istringstream line_stream(line);
        auto *new_rule = new Rule;
        line_stream >> new_rule->word >> new_rule->frequency;
        find_bigrams(new_rule->word, new_rule->bigrams);
        dict.insert(std::make_pair(new_rule->word,new_rule));
    }

    dict_file.close();
}

//clear_dict очищает память у словаря
//std::map<std::string, Rule *> &dict
void clear_dict(std::map<std::string, Rule *> &dict) {
    for (auto &rule : dict) {
        delete rule.second;
    }
}

//correct_word исправляет слово const std::string &word
//в соответсвии со словарем const std::map<std::string, Rule *> &dict
//возвращает правильное слово
std::string correct_word(const std::string &word,
                         const std::map<std::string, Rule *> &dict) {
    std::unordered_set<std::string> word_bigrams;
    find_bigrams(word, word_bigrams);

    Rule *correct_result = nullptr;
    std::pair<long, long>correct_proportion;
    for (auto &rule : dict) {
        long intersection_size = std::count_if(word_bigrams.begin(), word_bigrams.end(),
                [rule](std::string bigram) ->
                        bool { return rule.second->bigrams.find(bigram) != rule.second->bigrams.end();});
        long union_size = word_bigrams.size() + rule.second->bigrams.size() - intersection_size;

        if (correct_result == nullptr
            || intersection_size*correct_proportion.second > correct_proportion.first*union_size) {
            correct_result = rule.second;
            correct_proportion = std::make_pair(intersection_size, union_size);
        } else if (intersection_size*correct_proportion.second
                   == correct_proportion.first*union_size
                   && correct_result->frequency < rule.second->frequency) {
            correct_result = rule.second;
        }
    }

    return correct_result == nullptr ? "" : correct_result->word;
}

int main() {
    std::map<std::string, Rule *> dict;
    init_dict(dict_file_path, dict);
    std::string item;
    while (std::cin >> item) {
        std::cout << correct_word(item, dict) << std::endl;
    }
    clear_dict(dict);
    return 0;
}