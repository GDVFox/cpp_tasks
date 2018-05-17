//
// Created by gdvfox on 15.05.18.
//

#include <algorithm>
#include <iterator>

#include "textstats.hpp"

void get_tokens(const string &s,
                const unordered_set<char> &delimiters,
                vector<string> &tokens) {
    string item;
    for (char ch : s) {
        if (delimiters.find(ch) != delimiters.end()) {
            if (!item.empty()) {
                tokens.emplace_back(item);
                item.clear();
            }
        } else {
            item.push_back(tolower(ch));
        }
    }

    if (!item.empty()) tokens.emplace_back(item);
}

void get_type_freq(const vector<string> &tokens,
                   map<string, int> &freqdi) {
    for (const string &token : tokens) {
        auto it = freqdi.find(token);
        if (it != freqdi.end()) {
            it->second++;
        } else {
            freqdi[token] = 1;
        }
    }
}

void get_types(const vector<string> &tokens,
               vector<string> &wtypes) {
    copy(tokens.begin(), tokens.end(), back_inserter(wtypes));
    sort(wtypes.begin(), wtypes.end());
    wtypes.erase(unique(wtypes.begin(), wtypes.end()), wtypes.end());
}

void get_x_length_words(const vector<string> &wtypes,
                        int x, vector<string> &words) {
    copy_if(wtypes.begin(), wtypes.end(), back_inserter(words),
            [x](string word) -> bool { return word.length() >= x; });
}

void get_x_freq_words(const map<string, int> &freqdi,
                      int x, vector<string> &words) {
    //С transform было бы много накладных расходов
    for (auto it = freqdi.begin(); it != freqdi.end(); it++) {
        if (it->second >= x) words.push_back(it->first);
    }
}

void get_words_by_length_dict(const vector<string> &wtypes,
                              map<int, vector<string>> &lengthdi) {
    for (const string &token : wtypes) {
        auto it = lengthdi.find(token.length());
        if (it != lengthdi.end()) {
            it->second.push_back(token);
        } else {
            lengthdi[token.length()] = vector<string>({token});
        }
    }
}