use std::collections::HashMap;
use std::time::Instant;
use rand::prelude::*;

fn main() {
    let _currency = "PLN";
    let mut _balance = 100.0;
    
    let mut rarities: HashMap<&str, f32> = HashMap::new();
    let mut index = 0;

    if rarities.is_empty() {
        rarities.insert("♡", 0.9);
        rarities.insert("♢", 0.6);
        rarities.insert("♤", 0.4);
        rarities.insert("♧", 0.3);
        rarities.insert("♣", 0.2);
        rarities.insert("♦", 0.1);
        rarities.insert("♥", 0.05);
        rarities.insert("♠", 0.01);
        rarities.insert("✪", 0.003);
    }
    println!("\n\nStaring rarity roller...");
    let start_time = Instant::now();
    let mut rng = thread_rng(); // Deklaracja generatora na początku funkcji
    loop {
        index += 1;
        let mut new_char_list = Vec::new();
        for _ in 0..10 {
            let random_key = *rarities
            .iter()
            .max_by(|(_, &a), (_, &b)| a.partial_cmp(&b).unwrap())
            .map(|(key, _)| key)
            .unwrap();

            let chance = rng.gen::<f32>() * (0.9 - 0.003) + 0.003; // Losowanie z zakresu 0.0003 do 0.9
            let random_key = rarities
            .iter()
            .find(|(_, &value)| chance <= value)
            .map(|(key, _)| *key)
            .unwrap_or(random_key);
            new_char_list.push(random_key);
        }

        if new_char_list.contains(&"✪") {
            println!("\n\nWylosowano ✪ po {} próbach", index);
            break;
        } else {
            print!("\rPróba nr. {} z listą: {:?}            ",index,new_char_list);
            new_char_list.clear();
        }
    }
    let end_time = Instant::now();
    let duration = end_time - start_time;
    println!("Rolling took: {:?}\n\n", duration);
}
