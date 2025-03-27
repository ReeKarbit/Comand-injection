import java.io.IOException;
import java.util.Scanner;

public class Test2 {
    public static void main(String[] args) throws IOException {
        Scanner scanner = new Scanner(System.in);
        System.out.print("Masukkan nama file: ");
        String userInput = scanner.nextLine();

        // ❌ Rentan terhadap Command Injection
        ProcessBuilder pb = new ProcessBuilder("sh", "-c", "ls " + userInput);
        pb.start();
    }
}
